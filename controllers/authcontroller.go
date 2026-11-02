package controllers

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"real_world_backend/db"
	"real_world_backend/models"

	"real_world_backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoginInput struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	UserAgent string `json:"user_agent" binding:"required"`
}

func Login(c *gin.Context) {
	var input LoginInput
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
	tokenFamily := uuid.New().String()

	tokenRaw, newToken, err := services.NewRefreshToken(user.ID, tokenFamily, input.UserAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	if err := db.Create(newToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, services.TokenPair{
		AccessToken:  token,
		ExpeiresIn:   newToken.ExpiresAt.Unix(),
		RefreshToken: tokenRaw,
		TokenType:    "Bearer",
	})
}

type SignupInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
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

func Me(c *gin.Context) {
	id, found := c.Get("userID")
	if !found {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found"})
		return
	}

	conn := db.GetDB()
	var user models.User
	err := conn.WithContext(c).First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hi"})
		return
	}

	c.JSON(http.StatusOK, user)
}

type refreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func Refresh(c *gin.Context) {
	var input refreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn := db.GetDB()
	token, err := services.ConsumeRefreshToken(conn, services.HashToken(input.RefreshToken), time.Now())
	if err != nil {
		if errors.Is(err, services.ErrInvalidRefreshToken) {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, services.ErrRefreshTokenReused) {
			services.RevokeFamily(conn, token.FamilyID, time.Now())
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}
	}
	log.Print("consumed")
	tokenRaw, newToken, err := services.NewRefreshToken(token.UserID, token.FamilyID, token.UserAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	if err := conn.Create(newToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	var user models.User
	if err := conn.WithContext(c).Where("ID=?", newToken.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	log.Print(user)
	accessToken, err := services.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, services.TokenPair{
		AccessToken:  accessToken,
		ExpeiresIn:   newToken.ExpiresAt.Unix(),
		RefreshToken: tokenRaw,
		TokenType:    "Bearer",
	})
}

type LogoutInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func Logout(c *gin.Context) {
	var input LogoutInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	conn := db.GetDB()
	var token models.RefreshToken
	err := conn.Where("token_hash = ?", services.HashToken(input.RefreshToken)).First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := services.RevokeFamily(db.GetDB(), token.FamilyID, time.Now()); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
