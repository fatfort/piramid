package main

import (
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"

	"piramid/internal/config"
	"piramid/internal/database"
)

func main() {
	log.Println("Creating admin user...")

	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations first
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create production tenant
	tenant := &database.Tenant{
		Name:        "production-tenant",
		Description: "Production tenant for PIRAMID system",
		HomeNet:     cfg.HomeNet,
	}

	// Check if tenant already exists
	var existingTenant database.Tenant
	if err := db.Where("name = ?", tenant.Name).First(&existingTenant).Error; err != nil {
		// Tenant doesn't exist, create it
		if err := db.Create(tenant).Error; err != nil {
			log.Fatalf("Failed to create tenant: %v", err)
		}
		log.Printf("Created tenant: %s", tenant.Name)
	} else {
		tenant = &existingTenant
		log.Printf("Tenant already exists: %s", tenant.Name)
	}

	// Get admin credentials from environment variables
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@fatfort.com"
		log.Printf("ADMIN_EMAIL not set, using default: %s", adminEmail)
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "piramid-admin-2024"
		log.Printf("ADMIN_PASSWORD not set, using default password")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	user := &database.User{
		TenantID:     tenant.ID,
		Email:        adminEmail,
		PasswordHash: string(hashedPassword),
		Role:         "admin",
		Active:       true,
	}

	// Check if user already exists
	var existingUser database.User
	if err := db.Where("email = ?", user.Email).First(&existingUser).Error; err != nil {
		// User doesn't exist, create it
		if err := db.Create(user).Error; err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}
		log.Printf("Created user: %s", user.Email)
	} else {
		// User exists, update password and ensure active
		existingUser.PasswordHash = string(hashedPassword)
		existingUser.Role = "admin"
		existingUser.Active = true
		if err := db.Save(&existingUser).Error; err != nil {
			log.Fatalf("Failed to update user: %v", err)
		}
		log.Printf("Updated user: %s", user.Email)
	}

	log.Println("User creation completed successfully!")
	log.Printf("Login credentials: %s / %s", user.Email, adminPassword)
} 