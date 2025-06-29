package algo

import (
	"net"

	"github.com/oschwald/geoip2-golang"
)

// GeoIPDB wraps the MaxMind GeoIP database
type GeoIPDB struct {
	db *geoip2.Reader
}

// GeoLocation represents geographical information for an IP
type GeoLocation struct {
	Country   string  `json:"country"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	ASN       uint    `json:"asn,omitempty"`
	ISP       string  `json:"isp,omitempty"`
}

// NewGeoIPDB creates a new GeoIP database instance
func NewGeoIPDB(dbPath string) (*GeoIPDB, error) {
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, err
	}

	return &GeoIPDB{db: db}, nil
}

// Close closes the GeoIP database
func (g *GeoIPDB) Close() error {
	return g.db.Close()
}

// Lookup performs a GeoIP lookup for the given IP address
func (g *GeoIPDB) Lookup(ipStr string) (*GeoLocation, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return &GeoLocation{
			Country: "Unknown",
			City:    "Unknown",
		}, nil
	}

	record, err := g.db.City(ip)
	if err != nil {
		return &GeoLocation{
			Country: "Unknown",
			City:    "Unknown",
		}, nil
	}

	location := &GeoLocation{
		Country:   record.Country.Names["en"],
		City:      record.City.Names["en"],
		Latitude:  record.Location.Latitude,
		Longitude: record.Location.Longitude,
	}

	// Handle empty values
	if location.Country == "" {
		location.Country = "Unknown"
	}
	if location.City == "" {
		location.City = "Unknown"
	}

	return location, nil
}

// IsPrivateIP checks if an IP address is private
func IsPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Check for private IP ranges
	private := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"::1/128",
		"fc00::/7",
		"fe80::/10",
	}

	for _, cidr := range private {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// ValidateIP validates if a string is a valid IP address
func ValidateIP(ipStr string) bool {
	return net.ParseIP(ipStr) != nil
}
