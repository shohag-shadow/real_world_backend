package controllers

import (
	"net/http"
	"strings"

	"real_world_backend/db"
	"real_world_backend/models"

	"github.com/gin-gonic/gin"

	"real_world_backend/services"
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

	if !services.VerifyPassword(input.Password, passHash) {
		c.JSON(http.StatusBadRequest, gin.H{"messege": "passord is wrong"})
		return
	}
	token, tokenerr := services.GenerateToken(user)
	if tokenerr != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"messege": "error during creading token", "error": tokenerr.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"token": token})
}

func Signup(c *gin.Context) {
	var input SignupInput
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	passwordHash, hasherror := services.HashPassword(input.Password)
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
