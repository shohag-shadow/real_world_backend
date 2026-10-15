package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"real_world_backend/db"
	"real_world_backend/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/argon2"
)

type SignupInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func Login(c *gin.Context) {

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
