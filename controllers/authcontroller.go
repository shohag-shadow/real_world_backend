package controllers

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"real_world_backend/db"
	"real_world_backend/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
)

type SignupInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func Login(c *gin.Context) {
	var input LoginInput
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	db := db.GetDB()
	var user models.User
	var passHash = "$argon2id$v=19$m=65536$t=3$p=4$fvbE8VrW7tXvWCT4rZMpIw$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	dberr := db.WithContext(c).Where("email = ?", strings.ToLower(strings.TrimSpace(input.Email))).First(&user).Error
	if dberr == nil {
		passHash = user.PasswordHash
	}

	if !verifyPassword(input.Password, passHash) {
		c.JSON(http.StatusBadRequest, gin.H{"messege": "passord is wrong"})
		return
	}
	token, tokenerr := generateToken(user)
	if tokenerr != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"messege": "error during creading token", "error": tokenerr.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"token": token})
}

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func generateToken(user models.User) (string, error) {
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
func Signup(c *gin.Context) {
	var input SignupInput
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	passwordHash, hasherror := hashPassword(input.Password)
	if hasherror != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}
	user := models.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: passwordHash,
	}
	db := db.GetDB()
	if dbrestult := db.Create(&user); dbrestult.Error != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"messege": "User created successfully",
		"User":    user,
	})

}

const (
	memory      = 64 * 1024
	iterations  = 3
	parallelism = 4
	saltLength  = 16
	keyLength   = 32
)

func hashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)
	base64salt := base64.RawStdEncoding.EncodeToString(salt)
	base64hash := base64.RawStdEncoding.EncodeToString(hash)
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d$t=%d$p=%d$%s$%s", argon2.Version, memory, iterations, parallelism, base64salt, base64hash)
	return encodedHash, nil
}

func verifyPassword(password string, hash string) bool {
	parts := strings.Split(hash, "$")
	saltdecoded, err := base64.RawStdEncoding.DecodeString(parts[6])
	if err != nil {
		return false
	}
	hashdecoded, hasherr := base64.RawStdEncoding.DecodeString(parts[7])
	if hasherr != nil {
		return false
	}
	newHash := argon2.IDKey([]byte(password), saltdecoded, iterations, memory, parallelism, keyLength)
	return subtle.ConstantTimeCompare(newHash, hashdecoded) == 1
}
