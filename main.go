package main

import (
	"fmt"

	"real_world_backend/db"
	"real_world_backend/models"
)

func main() {

	db := db.GetDB()
	newUser := models.User{Name: "w3e63", Email: "w3e63@example.com"}
	result := db.Create(&newUser)
	if result.Error != nil {
		fmt.Printf("Notice: Record might already exist (%v)\n", result.Error)
	} else {
		fmt.Printf("Successfully inserted user with ID: %d\n", newUser.ID)
	}
}
