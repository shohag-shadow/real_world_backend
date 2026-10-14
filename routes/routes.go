package routes

import (
	"net/http"

	"real_world_backend/controllers"

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

	app.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
