package models

import (
	"time"

	"gorm.io/gorm"
)

// 1. Define your Database Model Struct
// GORM automatically mapping this to a "users" table
type User struct {
	ID              uint   `gorm:"primaryKey"`
	Name            string `gorm:"size:100;not null"`
	Email           string `gorm:"unique;not null"`
	PasswordHash    string `gorm:"size:256;not null"`
	EmailVarifiedAt time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}
