package main

import (
	"fmt"
	"inventory-system/internal/config"
	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	log.Println("Seeding Admin User...")

	db, err := config.NewDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Ensure schema exists
	if err := config.InitDB(db); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	userRepo := repository.NewUserRepository(db)

	username := "admin"
	password := "admin123"

	// Check if user already exists
	existing, err := userRepo.FindByUsername(username)
	if err != nil {
		log.Fatalf("Failed to check existing user: %v", err)
	}

	if existing != nil {
		log.Printf("User '%s' already exists. Skipping seed.", username)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	admin := &domain.User{
		Username: username,
		Password: string(hashedPassword),
		Role:     "admin",
		Status:   "active",
	}

	err = userRepo.Create(admin)
	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Println("--------------------------------------------------")
	fmt.Printf("Admin user created successfully!\n")
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Password: %s\n", password)
	fmt.Println("--------------------------------------------------")
}
