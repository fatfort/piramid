package algo

import (
	"encoding/json"
	"log"
	"strings"
	"time"
)

// SuricataEvent represents a Suricata eve.json event
type SuricataEvent struct {
	Timestamp string                 `json:"timestamp"`
	FlowID    uint64                 `json:"flow_id,omitempty"`
	InIface   string                 `json:"in_iface,omitempty"`
	EventType string                 `json:"event_type"`
	SrcIP     string                 `json:"src_ip"`
	SrcPort   int                    `json:"src_port,omitempty"`
	DestIP    string                 `json:"dest_ip"`
	DestPort  int                    `json:"dest_port,omitempty"`
	Proto     string                 `json:"proto,omitempty"`
	Alert     *Alert                 `json:"alert,omitempty"`
	SSH       *SSH                   `json:"ssh,omitempty"`
	HTTP      *HTTP                  `json:"http,omitempty"`
	DNS       *DNS                   `json:"dns,omitempty"`
	TLS       *TLS                   `json:"tls,omitempty"`
	Flow      *Flow                  `json:"flow,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Alert represents Suricata alert information
type Alert struct {
	Action      string                 `json:"action"`
	Gid         int                    `json:"gid"`
	SignatureID int                    `json:"signature_id"`
	Rev         int                    `json:"rev"`
	Signature   string                 `json:"signature"`
	Category    string                 `json:"category"`
	Severity    int                    `json:"severity"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SSH represents SSH-specific event data
type SSH struct {
	Client struct {
		ProtoVersion    string `json:"proto_version"`
		SoftwareVersion string `json:"software_version"`
	} `json:"client"`
	Server struct {
		ProtoVersion    string `json:"proto_version"`
		SoftwareVersion string `json:"software_version"`
	} `json:"server"`
}

// HTTP represents HTTP-specific event data
type HTTP struct {
	Hostname        string            `json:"hostname"`
	URL             string            `json:"url"`
	UserAgent       string            `json:"http_user_agent"`
	Method          string            `json:"http_method"`
	Protocol        string            `json:"protocol"`
	Status          int               `json:"status"`
	Length          int               `json:"length"`
	RequestHeaders  map[string]string `json:"request_headers,omitempty"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty"`
}

// DNS represents DNS-specific event data
type DNS struct {
	Type   string `json:"type"`
	Query  string `json:"query"`
	Answer string `json:"answer,omitempty"`
	Rcode  string `json:"rcode,omitempty"`
}

// TLS represents TLS-specific event data
type TLS struct {
	Subject   string `json:"subject"`
	Issuer    string `json:"issuer"`
	SNI       string `json:"sni,omitempty"`
	Version   string `json:"version,omitempty"`
	NotBefore string `json:"notbefore,omitempty"`
	NotAfter  string `json:"notafter,omitempty"`
}

// Flow represents flow information
type Flow struct {
	PktsToserver  int    `json:"pkts_toserver"`
	PktsToClient  int    `json:"pkts_toclient"`
	BytesToServer int    `json:"bytes_toserver"`
	BytesToClient int    `json:"bytes_toclient"`
	Start         string `json:"start"`
	End           string `json:"end,omitempty"`
	Age           int    `json:"age,omitempty"`
	State         string `json:"state,omitempty"`
	Reason        string `json:"reason,omitempty"`
	Alerted       bool   `json:"alerted,omitempty"`
}

// ParsedEvent represents a processed Suricata event with geo information
type ParsedEvent struct {
	ID         uint      `json:"id,omitempty"`
	TenantID   uint      `json:"tenant_id"`
	Timestamp  time.Time `json:"timestamp"`
	EventType  string    `json:"event_type"`
	SrcIP      string    `json:"src_ip"`
	SrcPort    int       `json:"src_port"`
	DestIP     string    `json:"dest_ip"`
	DestPort   int       `json:"dest_port"`
	Protocol   string    `json:"protocol"`
	Signature  string    `json:"signature"`
	Severity   int       `json:"severity"`
	Category   string    `json:"category"`
	Action     string    `json:"action"`
	Country    string    `json:"country"`
	City       string    `json:"city"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	RawPayload string    `json:"raw_payload"`
	CreatedAt  time.Time `json:"created_at"`
}

// Parser handles parsing of Suricata events
type Parser struct {
	geoip *GeoIPDB
}

// NewParser creates a new event parser
func NewParser(geoipDB *GeoIPDB) *Parser {
	return &Parser{
		geoip: geoipDB,
	}
}

// ParseEvent parses a Suricata eve.json event
func (p *Parser) ParseEvent(data []byte, tenantID uint) (*ParsedEvent, error) {
	var event SuricataEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}

	// Parse timestamp
	timestamp, err := time.Parse("2006-01-02T15:04:05.000000-0700", event.Timestamp)
	if err != nil {
		// Try alternative format
		timestamp, err = time.Parse("2006-01-02T15:04:05.000000Z", event.Timestamp)
		if err != nil {
			log.Printf("Failed to parse timestamp: %v", err)
			timestamp = time.Now()
		}
	}

	// Create parsed event
	parsed := &ParsedEvent{
		TenantID:   tenantID,
		Timestamp:  timestamp,
		EventType:  event.EventType,
		SrcIP:      event.SrcIP,
		SrcPort:    event.SrcPort,
		DestIP:     event.DestIP,
		DestPort:   event.DestPort,
		Protocol:   event.Proto,
		RawPayload: string(data),
		CreatedAt:  time.Now(),
	}

	// Add alert-specific information
	if event.Alert != nil {
		parsed.Signature = event.Alert.Signature
		parsed.Severity = event.Alert.Severity
		parsed.Category = event.Alert.Category
		parsed.Action = event.Alert.Action
	}

	// Perform GeoIP lookup for source IP
	if p.geoip != nil && !IsPrivateIP(event.SrcIP) {
		if geo, err := p.geoip.Lookup(event.SrcIP); err == nil {
			parsed.Country = geo.Country
			parsed.City = geo.City
			parsed.Latitude = geo.Latitude
			parsed.Longitude = geo.Longitude
		}
	}

	return parsed, nil
}

// IsSSHBruteForce checks if an event indicates SSH brute force
func (p *Parser) IsSSHBruteForce(event *SuricataEvent) bool {
	if event.EventType != "alert" || event.Alert == nil {
		return false
	}

	// Check for SSH-related signatures that indicate brute force
	signature := event.Alert.Signature
	return contains(signature, []string{
		"SSH", "brute", "login", "authentication", "failed",
	})
}

// contains checks if any of the keywords exist in the text (case-insensitive)
func contains(text string, keywords []string) bool {
	text = strings.ToLower(text)
	for _, keyword := range keywords {
		if strings.Contains(text, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// GetEventPriority returns the priority level of an event
func (p *Parser) GetEventPriority(event *SuricataEvent) int {
	if event.Alert != nil {
		// Suricata severity: 1 = high, 2 = medium, 3 = low, 4 = info
		return event.Alert.Severity
	}
	return 3 // Default to low priority
}

// ExtractIOCs extracts Indicators of Compromise from an event
func (p *Parser) ExtractIOCs(event *SuricataEvent) map[string][]string {
	iocs := make(map[string][]string)

	// Extract IPs
	if ValidateIP(event.SrcIP) {
		iocs["ip"] = append(iocs["ip"], event.SrcIP)
	}
	if ValidateIP(event.DestIP) {
		iocs["ip"] = append(iocs["ip"], event.DestIP)
	}

	// Extract domains from HTTP events
	if event.HTTP != nil && event.HTTP.Hostname != "" {
		iocs["domain"] = append(iocs["domain"], event.HTTP.Hostname)
	}

	// Extract domains from DNS events
	if event.DNS != nil && event.DNS.Query != "" {
		iocs["domain"] = append(iocs["domain"], event.DNS.Query)
	}

	return iocs
}
