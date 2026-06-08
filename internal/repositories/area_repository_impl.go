package repositories

import (
	"jejak/internal/models"

	"gorm.io/gorm"
)

type AreaRepositoryImpl struct {
	db *gorm.DB
}

func NewAreaRepository(db *gorm.DB) AreaRepository {
	return &AreaRepositoryImpl{db: db}
}

func (r *AreaRepositoryImpl) Create(area *models.Area) error {
	return r.db.Create(area).Error
}

func (r *AreaRepositoryImpl) FindAll() ([]models.Area, error) {
	var areas []models.Area
	if err := r.db.Find(&areas).Error; err != nil {
		return nil, err
	}
	return areas, nil
}

func (r *AreaRepositoryImpl) FindByID(id string) (*models.Area, error) {
	var area models.Area
	if err := r.db.Where("id = ?", id).First(&area).Error; err != nil {
		return nil, err
	}
	return &area, nil
}

func (r *AreaRepositoryImpl) Update(area *models.Area) error {
	result := r.db.Model(&models.Area{}).
		Where("id = ?", area.ID).
		Updates(area)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *AreaRepositoryImpl) Delete(id string) error {
	result := r.db.Where("id = ?", id).Delete(&models.Area{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *AreaRepositoryImpl) FindByName(name string) (*models.Area, error) {
	var area models.Area
	if err := r.db.Where("name = ?", name).First(&area).Error; err != nil {
		return nil, err
	}
	return &area, nil
}
