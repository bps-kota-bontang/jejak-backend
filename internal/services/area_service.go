package services

import (
	"errors"
	"net/http"
	"strings"

	"jejak/internal/dto"
	apperrors "jejak/internal/errors"
	"jejak/internal/models"
	"jejak/internal/repositories"

	"gorm.io/gorm"
)

type AreaService struct {
	areaRepo repositories.AreaRepository
}

func NewAreaService(areaRepo repositories.AreaRepository) *AreaService {
	return &AreaService{
		areaRepo: areaRepo,
	}
}

func (s *AreaService) CreateArea(req dto.CreateAreaRequest) error {
	name := strings.TrimSpace(req.Name)
	geoJSONFilePath := strings.TrimSpace(req.GeoJSONFilePath)
	listKeys := sanitizeListKeys(req.ListKeys)

	if name == "" {
		return apperrors.NewHttpError(http.StatusBadRequest, "name is required")
	}

	if geoJSONFilePath == "" {
		return apperrors.NewHttpError(http.StatusBadRequest, "geojson_file_path is required")
	}

	if len(listKeys) == 0 {
		return apperrors.NewHttpError(http.StatusBadRequest, "list_keys is required")
	}

	area := &models.Area{
		Name:            name,
		GeoJSONFilePath: geoJSONFilePath,
		ListKeys:        listKeys,
		Description:     req.Description,
	}

	return s.areaRepo.Create(area)
}

func (s *AreaService) GetAllAreas() ([]models.Area, error) {
	return s.areaRepo.FindAll()
}

func (s *AreaService) GetAreaByID(id string) (*models.Area, error) {
	if id == "" {
		return nil, apperrors.NewHttpError(http.StatusBadRequest, "id is required")
	}

	area, err := s.areaRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewHttpError(http.StatusNotFound, "Area not found")
		}
		return nil, err
	}

	return area, nil
}

func (s *AreaService) UpdateArea(id string, req dto.UpdateAreaRequest) error {
	if id == "" {
		return apperrors.NewHttpError(http.StatusBadRequest, "id is required")
	}

	area, err := s.areaRepo.FindByID(id)
	if err != nil {
		return err
	}

	if req.Name != "" {
		area.Name = strings.TrimSpace(req.Name)
	}

	if req.GeoJSONFilePath != "" {
		area.GeoJSONFilePath = strings.TrimSpace(req.GeoJSONFilePath)
	}

	if req.ListKeys != nil {
		area.ListKeys = sanitizeListKeys(req.ListKeys)
	}

	if req.Description != nil {
		area.Description = req.Description
	}

	return s.areaRepo.Update(area)
}

func (s *AreaService) DeleteArea(id string) error {
	if id == "" {
		return apperrors.NewHttpError(http.StatusBadRequest, "id is required")
	}

	return s.areaRepo.Delete(id)
}

func (s *AreaService) GetAreaByName(name string) (*models.Area, error) {
	if name == "" {
		return nil, apperrors.NewHttpError(http.StatusBadRequest, "name is required")
	}

	return s.areaRepo.FindByName(name)
}

func sanitizeListKeys(keys []string) []string {
	if len(keys) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(keys))
	result := make([]string, 0, len(keys))
	for _, item := range keys {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}

	return result
}
