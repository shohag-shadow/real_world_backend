package main

import (
	"log"

	"real_world_backend/db"

	"real_world_backend/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading environment using proper machiene environment")
	}
	db := db.GetDB()
	if db != nil {
		log.Println("Connected to database and working")
	}
	app := gin.Default()
	routes.RegisterRoutes(app)
	app.Run()
}
