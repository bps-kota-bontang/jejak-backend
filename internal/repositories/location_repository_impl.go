package repositories

import (
	"jejak/internal/models"

	"gorm.io/gorm"
)

type LocationRepositoryImpl struct {
	db *gorm.DB
}

func NewLocationRepository(db *gorm.DB) LocationRepository {
	return &LocationRepositoryImpl{db: db}
}

func (r *LocationRepositoryImpl) ReplaceByAssignmentID(assignmentID string, locations []models.Location) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("assignment_id = ?", assignmentID).Delete(&models.Location{}).Error; err != nil {
			return err
		}
		if len(locations) == 0 {
			return nil
		}
		return tx.Create(&locations).Error
	})
}

func (r *LocationRepositoryImpl) FindByAssignmentID(assignmentID string) ([]models.Location, error) {
	var locations []models.Location
	if err := r.db.Where("assignment_id = ?", assignmentID).Find(&locations).Error; err != nil {
		return nil, err
	}
	return locations, nil
}
