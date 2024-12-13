package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/oadultradeepfield/olympliance-server/internal/databases"
	"github.com/oadultradeepfield/olympliance-server/internal/routes"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	if os.Getenv("GO_ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	db := databases.InitDB()

	r := gin.New()

	routes.InitRoutes(r, db)

	if err := r.Run(":" + os.Getenv("PORT")); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
