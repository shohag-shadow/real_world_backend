package db

import (
	"log"
	"os"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once
)

func GetDB() *gorm.DB {
	once.Do(func() {
		dbUrl := os.Getenv("DATABASE_URL")
		if dbUrl == "" {
			log.Fatal("Database url not found")
		}
		var err error
		db, err = gorm.Open(postgres.Open(dbUrl))
		if err != nil {
			log.Fatal(dbUrl)
			log.Fatal("Database loading failed")
		}
	})
	return db
}
