package services

import (
	"os"
	"real_world_backend/models"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateToken(user models.User) (string, error) {
	expirationTime := time.Now().Add(100 * time.Second)
	claims := &Claims{
		UserID: uint(user.ID),
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	print(os.Getenv("JWT_SECREAT"))
	return token.SignedString([]byte(os.Getenv("JWT_SECREAT")))
}
