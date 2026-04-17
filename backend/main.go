package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"test-app/backend/config"
	"test-app/backend/models"
	"test-app/backend/routes"
)

func main() {
	cfg := config.Load()

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	var db *gorm.DB
	var err error
	for i := 0; i < 5; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			log.Println("Connected to database")
			break
		}
		log.Printf("Failed to connect to database (attempt %d/5): %v", i+1, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatalf("Could not connect to database after 5 attempts: %v", err)
	}

	// Auto-migrate
	if err := db.AutoMigrate(&models.AdminUser{}, &models.Profile{}, &models.Order{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migrated")

	// Seed admin user
	seedAdmin(db, cfg)

	// Seed sample data
	seedSampleData(db)

	// Setup routes and start server
	r := routes.Setup(db, cfg.JWTSecret)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func seedAdmin(db *gorm.DB, cfg *config.Config) {
	var count int64
	db.Model(&models.AdminUser{}).Where("username = ?", cfg.AdminUsername).Count(&count)
	if count > 0 {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash admin password: %v", err)
	}

	admin := models.AdminUser{
		Username:     cfg.AdminUsername,
		PasswordHash: string(hash),
	}
	if err := db.Create(&admin).Error; err != nil {
		log.Fatalf("Failed to seed admin user: %v", err)
	}
	log.Println("Admin user seeded")
}

func seedSampleData(db *gorm.DB) {
	var count int64
	db.Model(&models.Profile{}).Count(&count)
	if count > 0 {
		return
	}

	hashPassword := func(password string) string {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}
		return string(hash)
	}

	now := time.Now()
	deadline := now.AddDate(0, 1, 0)

	// Profile+ 1
	plus1 := models.Profile{
		Type:            "plus",
		Name:            "Plus Company A",
		Email:           "plus1@example.com",
		PasswordHash:    hashPassword("plus123"),
		Commission:      10.0,
		PaymentDeadline: &deadline,
	}
	db.Create(&plus1)

	// Profile+ 2
	plus2 := models.Profile{
		Type:            "plus",
		Name:            "Plus Company B",
		Email:           "plus2@example.com",
		PasswordHash:    hashPassword("plus123"),
		Commission:      12.5,
		PaymentDeadline: &deadline,
	}
	db.Create(&plus2)

	// Standard 1 under Plus 1
	std1 := models.Profile{
		Type:            "standard",
		ParentID:        &plus1.ID,
		Name:            "Standard Partner 1",
		Email:           "std1@example.com",
		PasswordHash:    hashPassword("std123"),
		Commission:      7.0,
		PaymentDeadline: &deadline,
	}
	db.Create(&std1)

	// Standard 2 under Plus 1
	std2 := models.Profile{
		Type:            "standard",
		ParentID:        &plus1.ID,
		Name:            "Standard Partner 2",
		Email:           "std2@example.com",
		PasswordHash:    hashPassword("std123"),
		Commission:      8.0,
		PaymentDeadline: &deadline,
	}
	db.Create(&std2)

	// Min 1 under Standard 1
	min1 := models.Profile{
		Type:         "min",
		ParentID:     &std1.ID,
		Name:         "Min Agent 1",
		Email:        "min1@example.com",
		PasswordHash: hashPassword("min123"),
		Commission:   3.0,
	}
	db.Create(&min1)

	// Min 2 under Standard 1
	min2 := models.Profile{
		Type:         "min",
		ParentID:     &std1.ID,
		Name:         "Min Agent 2",
		Email:        "min2@example.com",
		PasswordHash: hashPassword("min123"),
		Commission:   4.0,
	}
	db.Create(&min2)

	// Orders
	orders := []models.Order{
		{ProfileID: plus1.ID, Number: "ORD-001", Amount: 50000, Paid: 50000, CreatedAt: now},
		{ProfileID: plus1.ID, Number: "ORD-002", Amount: 75000, Paid: 25000, CreatedAt: now},
		{ProfileID: plus2.ID, Number: "ORD-003", Amount: 30000, Paid: 0, CreatedAt: now},
		{ProfileID: std1.ID, Number: "ORD-004", Amount: 20000, Paid: 20000, CreatedAt: now},
		{ProfileID: std1.ID, Number: "ORD-005", Amount: 15000, Paid: 5000, CreatedAt: now},
		{ProfileID: std2.ID, Number: "ORD-006", Amount: 10000, Paid: 0, CreatedAt: now},
		{ProfileID: min1.ID, Number: "ORD-007", Amount: 8000, Paid: 8000, CreatedAt: now},
		{ProfileID: min2.ID, Number: "ORD-008", Amount: 5000, Paid: 2000, CreatedAt: now},
	}

	for _, order := range orders {
		db.Create(&order)
	}

	log.Println("Sample data seeded")
}
