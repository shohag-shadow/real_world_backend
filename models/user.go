package models

import (
	"time"
)

type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"not null" json:"name"`
	Email        string    `gorm:"uniqueIndex:user_email_unique;not null" json:"email"`
	PasswordHash string    `gorm:"column:password_hash;not null" json:"-"`
	CreatedAt    time.Time `gorm:"default:now();not null" json:"-"`
	UpdatedAt    time.Time `gorm:"default:now();not null" json:"-"`
}
