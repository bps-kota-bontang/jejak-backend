package repositories

import "jejak/internal/models"

type AreaRepository interface {
	Create(area *models.Area) error
	FindAll() ([]models.Area, error)
	FindByID(id string) (*models.Area, error)
	Update(area *models.Area) error
	Delete(id string) error
	FindByName(name string) (*models.Area, error)
}
