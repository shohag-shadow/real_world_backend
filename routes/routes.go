package routes

import (
	"net/http"

	"real_world_backend/controllers"
	"real_world_backend/middlewires"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(app *gin.Engine) {

	api := app.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/signup", controllers.Signup)
			auth.POST("/login", controllers.Login)
		}
	}
	api.GET("/me", middlewires.AuthRequired(), controllers.Me)

	app.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
