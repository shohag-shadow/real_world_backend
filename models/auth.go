package models

import "time"

type RefreshToken struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    string    `gorm:"not null;index" json:"user_id"`
	TokenHash string    `gorm:"not null;uniqueIndex" json:"-"`
	FamilyID  string    `gorm:"not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	UsedAt    time.Time `json:"-"`
	RevokedAT time.Time `json:"-"`
	UserAgent string    `json:"-"`
	CreatedAt time.Time `json:"-"`
}

type PasswordRefreshToken struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserId     string    `gorm:"not null;index" json:"user_id"`
	TokenHash  string    `gorm:"not null;uniueIndex" json:"-"`
	ExpiresAt  time.Time `gorm:"not null" json:"expires_at"`
	ConsumedAt time.Time `json:"-"`
	CreatedAt  time.Time `json:"-"`
}
