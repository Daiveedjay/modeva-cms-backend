package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/services"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// init loads environment variables
func init() {
	_ = godotenv.Load()
}

// main creates a super admin account
// Usage: go run cmd/seed/main.go
// This is a standalone CLI tool, not part of the main application
func main() {
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("MODEVA CMS - Super Admin Seeder")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	// Initialize database connections
	config.InitDB()
	log.Println("✓ Connected to databases")

	// Get input from user
	email, password, name := getAdminCredentials()

	// Check if admin already exists
	var existingAdmin models.Admin
	if err := config.CmsGorm.Where("email = ?", email).First(&existingAdmin).Error; err == nil {
		fmt.Printf("❌ Admin with email '%s' already exists\n", email)
		os.Exit(1)
	} else if err != gorm.ErrRecordNotFound {
		log.Fatalf("Database error: %v", err)
	}
	log.Printf("✓ Email '%s' is available", email)

	// Hash password
	authService := services.GetAdminAuthService()
	passwordHash, err := authService.HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	log.Println("✓ Password hashed securely")

	// Create super admin
	superAdmin := models.Admin{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
		Role:         "super_admin",
		Status:       "active",
	}

	if err := config.CmsGorm.Create(&superAdmin).Error; err != nil {
		log.Fatalf("Failed to create super admin: %v", err)
	}

	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("✅ Super Admin Created Successfully!")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Printf("ID:    %s\n", superAdmin.ID)
	fmt.Printf("Email: %s\n", superAdmin.Email)
	fmt.Printf("Name:  %s\n", superAdmin.Name)
	fmt.Printf("Role:  %s\n", superAdmin.Role)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Start the CMS server: go run main.go")
	fmt.Println("2. Login at POST /api/v1/admin/login with email and password")
	fmt.Println("3. Use the returned token for authenticated requests")
	fmt.Println("4. Invite other admins using POST /api/v1/admin/invites")
	fmt.Println()
}

// getAdminCredentials prompts user for admin details
func getAdminCredentials() (email, password, name string) {
	fmt.Println("Enter Super Admin Details:")
	fmt.Println()

	// Email
	for {
		fmt.Print("Email: ")
		fmt.Scanln(&email)
		if email != "" {
			break
		}
		fmt.Println("❌ Email cannot be empty")
	}

	// Name
	for {
		fmt.Print("Name: ")
		fmt.Scanln(&name)
		if name != "" {
			break
		}
		fmt.Println("❌ Name cannot be empty")
	}

	// Password
	for {
		fmt.Print("Password (min 8 characters): ")
		fmt.Scanln(&password)

		authService := services.GetAdminAuthService()
		if !authService.ValidatePassword(password) {
			fmt.Println("❌ Password must be at least 8 characters")
			continue
		}
		break
	}

	// Confirm password
	for {
		fmt.Print("Confirm Password: ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm == password {
			break
		}
		fmt.Println("❌ Passwords do not match")
	}

	fmt.Println()
	return email, password, name
}
