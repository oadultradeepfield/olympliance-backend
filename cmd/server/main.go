package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/oadultradeepfield/olympliance-server/internal/databases"
	"github.com/oadultradeepfield/olympliance-server/internal/routes"
	"github.com/oadultradeepfield/olympliance-server/internal/services"
)

func main() {
	if os.Getenv("GO_ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	db := databases.InitDB()

	reputationCalculator := services.NewReputationCalculator(db)
	reputationCalculator.CalculateReputationOnStartup()

	r := gin.Default()

	routes.InitRoutes(r, db)

	if err := r.Run(":" + os.Getenv("PORT")); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
