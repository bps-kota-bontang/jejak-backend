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

	// Backfill event_hash for existing logs with NULL or empty hash, and cleanup duplicates
	if err := backfillLogEventHashes(db); err != nil {
		log.Printf("Warning: failed to backfill log event hashes: %v", err)
	}

	log.Println("Database connection established successfully")
	return db, nil
}

func backfillLogEventHashes(db *gorm.DB) error {
	var logsWithoutHash []models.Log
	if err := db.Where("event_hash IS NULL OR event_hash = ''").Order("assignment_id, created_at ASC").Find(&logsWithoutHash).Error; err != nil {
		return err
	}

	if len(logsWithoutHash) == 0 {
		return nil
	}

	logsByAssignment := make(map[string][]models.Log)
	for _, item := range logsWithoutHash {
		logsByAssignment[item.AssignmentID] = append(logsByAssignment[item.AssignmentID], item)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for assignmentID, logs := range logsByAssignment {
			seenHashes := make(map[string]string)
			for _, item := range logs {
				hash := models.BuildLogEventHash(item.AssignmentID, item.Action, item.Latitude, item.Longitude, item.ActionedAt)

				if _, exists := seenHashes[hash]; exists {
					continue
				}

				seenHashes[hash] = item.ID
				if err := tx.Model(&models.Log{}).Where("id = ?", item.ID).Update("event_hash", hash).Error; err != nil {
					return err
				}
			}

			log.Printf("Backfilled event_hash for %d logs in assignment %s", len(seenHashes), assignmentID)
		}
		return nil
	})
}
