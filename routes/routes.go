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
		api.POST("/signup", controllers.Signup)
		api.POST("/login", controllers.Login)
		api.POST("/refresh", controllers.Refresh)
		api.POST("/logout", controllers.Logout)
	}
	api.GET("/me", middlewires.AuthRequired(), controllers.Me)

	app.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
