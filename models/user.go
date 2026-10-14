package models

import (
	"time"
)

type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"`
	Name         string    `gorm:"not null"`
	Email        string    `gorm:"uniqueIndex:user_email_unique;not null"`
	PasswordHash string    `gorm:"column:password_hash;not null"`
	CreatedAt    time.Time `gorm:"default:now();not null"`
	UpdatedAt    time.Time `gorm:"default:now();not null"`
}
