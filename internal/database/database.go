package database

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect establishes a connection to the PostgreSQL database
func Connect(databaseURL string) (*gorm.DB, error) {
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(databaseURL), config)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// Migrate runs database migrations
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Tenant{},
		&User{},
		&Event{},
		&IPBan{},
	)
}

// Models

// Tenant represents a tenant in the system
type Tenant struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"unique;not null" json:"name"`
	Description string    `json:"description"`
	HomeNet     string    `json:"home_net"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Users  []User  `gorm:"foreignKey:TenantID" json:"users,omitempty"`
	Events []Event `gorm:"foreignKey:TenantID" json:"events,omitempty"`
}

// User represents a user in the system
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	TenantID     uint      `gorm:"not null" json:"tenant_id"`
	Email        string    `gorm:"unique;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Role         string    `gorm:"default:'user'" json:"role"`
	Active       bool      `gorm:"default:true" json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

// Event represents a security event from Suricata
type Event struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TenantID   uint      `gorm:"not null" json:"tenant_id"`
	Timestamp  time.Time `gorm:"not null" json:"timestamp"`
	EventType  string    `gorm:"not null" json:"event_type"`
	SrcIP      string    `gorm:"not null;index" json:"src_ip"`
	SrcPort    int       `json:"src_port"`
	DestIP     string    `gorm:"not null;index" json:"dest_ip"`
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
	RawPayload string    `gorm:"type:text" json:"raw_payload"`
	CreatedAt  time.Time `json:"created_at"`

	// Relationships
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

// IPBan represents a banned IP address
type IPBan struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	TenantID  uint       `gorm:"not null" json:"tenant_id"`
	IP        string     `gorm:"not null;index" json:"ip"`
	Reason    string     `json:"reason"`
	BannedBy  uint       `json:"banned_by"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`

	// Relationships
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	User   User   `gorm:"foreignKey:BannedBy" json:"banned_by_user,omitempty"`
}
