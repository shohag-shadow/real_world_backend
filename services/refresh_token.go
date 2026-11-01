package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"real_world_backend/models"
	"time"

	"gorm.io/gorm"
)

const defaultRefreshTokenTTL = 30 * 24 * time.Hour

var (
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenReused  = errors.New("refresh token reuse detected")
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpeiresIn   int64  `json:"expeires_in"`
}

func refreshTokenTTL() time.Duration {
	if raw := os.Getenv("REFRESH_TOKEN_TTL"); raw != "" {
		if ttl, err := time.ParseDuration(raw); err == nil && ttl > 0 {
			return ttl
		}
	}
	return defaultRefreshTokenTTL
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func RevokeFamily(conn *gorm.DB, familyID string, at time.Time) error {
	return conn.Model(&models.RefreshToken{}).
		Where("family_id = ? AND revoked_at IS NULL", familyID).
		Update("revoked_at", at).Error
}

func RevokeAllUserTokens(conn *gorm.DB, userID uint64, at time.Time) error {
	return conn.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", at).Error
}

func NewRefreshToken(userId uint64, familyId string, userAgent string) (string, *models.RefreshToken, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", nil, err
	}
	raw := base64.RawURLEncoding.EncodeToString(b)
	now := time.Now()
	token := models.RefreshToken{
		UserID:    userId,
		TokenHash: HashToken(raw),
		FamilyID:  familyId,
		ExpiresAt: now.Add(refreshTokenTTL()),
		UserAgent: userAgent,
	}
	return raw, &token, nil
}

func ConsumeRefreshToken(tx *gorm.DB, hashedToken string, now time.Time) (*models.RefreshToken, error) {
	var presented models.RefreshToken
	err := tx.Where("token_hash = ?", hashedToken).First(&presented).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidRefreshToken
	}
	if err != nil {
		return nil, err
	}
	if presented.UsedAt != nil {
		log.Printf("refresh token reuse detected for family %s; revoking the family", presented.FamilyID)
		return nil, ErrRefreshTokenReused
	}
	if presented.RevokedAt != nil || !now.Before(presented.ExpiresAt) {
		return nil, ErrInvalidRefreshToken
	}
	err = tx.Transaction(func(innerTx *gorm.DB) error {
		result := innerTx.Model(&models.RefreshToken{}).
			Where("id = ? AND used_at IS NULL AND revoked_at IS NULL", presented.ID).
			Update("used_at", now)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrRefreshTokenReused
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &presented, nil
}
