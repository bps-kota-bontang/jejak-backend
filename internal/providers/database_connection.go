package providers

import (
	"fmt"
	"jejak/config"
	"jejak/internal/models"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewDBConnection establishes a connection to the database based on provided config and returns the DB instance
func NewDBConnection(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// Initialize the DSN (Data Source Name) and Dialector based on the DB driver
	var dsn string
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "postgres":
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port)
		dialector = postgres.Open(dsn)

	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
		dialector = mysql.Open(dsn)

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	// Open the DB connection
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
		return nil, err
	}

	// Run database migrations
	if err := db.AutoMigrate(
		&models.User{},
		&models.Survey{},
		&models.Region{},
		&models.Assignment{},
		&models.Log{},
		&models.Answer{},
		&models.Location{},
		&models.Area{},
	); err != nil {
		log.Fatalf("Failed to apply database migrations: %v", err)
		return nil, err
	}

	log.Println("Database connection established successfully")
	return db, nil
}
