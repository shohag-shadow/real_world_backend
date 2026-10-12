package db

import (
	"fmt"
	"log"
	"real_world_backend/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func GetDB() *gorm.DB {
	dbFile := "db/gorm.db"

	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	fmt.Println("Successfully connected to SQLite using GORM!")

	// 3. Run the AutoMigration
	// GORM will safely create tables, missing columns, and indexes.
	// It will NOT drop or alter existing columns to protect data.
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	fmt.Println("Database migration completed successfully!")
	return db
}
