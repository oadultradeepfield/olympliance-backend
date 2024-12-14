package databases

import (
	"log"
	"os"

	"github.com/oadultradeepfield/olympliance-server/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() *gorm.DB {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		log.Fatalf("DSN is not set in the environment variables")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Info),
		PrepareStmt: true,
	})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Thread{},
		&models.Comment{},
		&models.Interaction{},
		&models.Category{},
	); err != nil {
		log.Fatalf("Error migrating database: %v", err)
	}

	categories := []models.Category{
		{Name: "General"},
		{Name: "Mathematics"},
		{Name: "Physics"},
		{Name: "Chemistry"},
		{Name: "Informatics"},
		{Name: "Biology"},
		{Name: "Philosophy"},
		{Name: "Astronomy"},
		{Name: "Geography"},
		{Name: "Linguistics"},
		{Name: "Earth Science"},
	}

	for _, category := range categories {
		result := db.FirstOrCreate(&category, models.Category{Name: category.Name})
		if result.Error != nil {
			log.Fatalf("Error creating category %s: %v\n", category.Name, result.Error)
		} else {
			log.Printf("Category '%s' checked/added successfully!\n", category.Name)
		}
	}

	log.Println("Database connected, migrated, and categories added successfully")

	return db
}
