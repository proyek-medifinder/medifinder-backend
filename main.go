package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "github.com/sasaefulanwar/medifinder/docs"
	"github.com/sasaefulanwar/medifinder/internal/config"
	"github.com/sasaefulanwar/medifinder/internal/routes"
)

// @title Medifinder API
// @version 1.0
// @description API Server for Medifinder Application
// @host localhost:8080
// @BasePath
func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db := config.ConnectDB()
	defer db.Close()

	r := routes.SetupRouter(db)

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("APP_PORT") 
	}
	if port == "" {
		port = "8080" 
	}
	r.Run(":" + port)
}
