package main

import (
	"log"

	"golang.org/x/crypto/bcrypt"

	"piramid/internal/config"
	"piramid/internal/database"
)

func main() {
	log.Println("Starting database seeding...")

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

	// Create default tenant
	tenant := &database.Tenant{
		Name:        "fatfort-internal",
		Description: "Default internal tenant for PIRAMID system",
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

	// Create default admin user
	password := "admin123" // In production, this should be configurable
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	user := &database.User{
		TenantID:     tenant.ID,
		Email:        "admin@fatfort.local",
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
		log.Printf("Created admin user: %s (password: %s)", user.Email, password)
	} else {
		log.Printf("Admin user already exists: %s", user.Email)
	}

	// Create a regular user as well
	regularPassword := "user123"
	hashedRegularPassword, err := bcrypt.GenerateFromPassword([]byte(regularPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash regular user password: %v", err)
	}

	regularUser := &database.User{
		TenantID:     tenant.ID,
		Email:        "user@fatfort.local",
		PasswordHash: string(hashedRegularPassword),
		Role:         "user",
		Active:       true,
	}

	var existingRegularUser database.User
	if err := db.Where("email = ?", regularUser.Email).First(&existingRegularUser).Error; err != nil {
		if err := db.Create(regularUser).Error; err != nil {
			log.Fatalf("Failed to create regular user: %v", err)
		}
		log.Printf("Created regular user: %s (password: %s)", regularUser.Email, regularPassword)
	} else {
		log.Printf("Regular user already exists: %s", regularUser.Email)
	}

	log.Println("Database seeding completed successfully!")
	log.Println("You can now login with:")
	log.Printf("  Admin: admin@fatfort.local / admin123")
	log.Printf("  User:  user@fatfort.local / user123")
}
