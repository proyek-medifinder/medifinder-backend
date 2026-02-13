package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/sasaefulanwar/medifinder/internal/config"
	"github.com/sasaefulanwar/medifinder/internal/routes"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db := config.ConnectDB()
	defer db.Close()

	r := routes.SetupRouter(db)
	
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	port := os.Getenv("APP_PORT")
	r.Run(":" + port)
}
