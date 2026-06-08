package handlers

import (
	"fmt"
	"io"
	"jejak/internal/dto"
	apperrors "jejak/internal/errors"
	"jejak/internal/services"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

var invalidFileNameChars = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

type AreaHandler struct {
	service  *services.AreaService
	validate *validator.Validate
}

func NewAreaHandler(service *services.AreaService, validate *validator.Validate) *AreaHandler {
	return &AreaHandler{
		service:  service,
		validate: validate,
	}
}

func (h *AreaHandler) GetAll(c fiber.Ctx) error {
	areas, err := h.service.GetAllAreas()
	if err != nil {
		return respondError(c, err)
	}

	responses := make([]dto.AreaResponse, len(areas))
	for i, area := range areas {
		responses[i] = dto.AreaResponse{
			ID:              area.ID,
			Name:            area.Name,
			GeoJSONFilePath: area.GeoJSONFilePath,
			ListKeys:        area.ListKeys,
			Description:     area.Description,
			CreatedAt:       area.CreatedAt,
			UpdatedAt:       area.UpdatedAt,
		}
	}

	return respondOK(c, responses, "Areas retrieved successfully")
}

func (h *AreaHandler) GetByID(c fiber.Ctx) error {
	id := c.Params("id")

	area, err := h.service.GetAreaByID(id)
	if err != nil {
		return respondError(c, err)
	}

	response := dto.AreaResponse{
		ID:              area.ID,
		Name:            area.Name,
		GeoJSONFilePath: area.GeoJSONFilePath,
		ListKeys:        area.ListKeys,
		Description:     area.Description,
		CreatedAt:       area.CreatedAt,
		UpdatedAt:       area.UpdatedAt,
	}

	return respondOK(c, response, "Area retrieved successfully")
}

func (h *AreaHandler) Create(c fiber.Ctx) error {
	var req dto.CreateAreaRequest
	if err := c.Bind().Body(&req); err != nil {
		return respondValidation(c, err)
	}

	if err := h.validate.Struct(req); err != nil {
		return respondValidation(c, err)
	}

	if err := h.service.CreateArea(req); err != nil {
		return respondError(c, err)
	}

	return respondCreated(c, nil, "Area created successfully")
}

func (h *AreaHandler) UploadGeoJSON(c fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "file is required"))
	}

	if fileHeader.Size <= 0 {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "file is empty"))
	}

	if fileHeader.Size > 20*1024*1024 {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "file size exceeds 20MB limit"))
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".geojson" && ext != ".json" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "only .geojson or .json files are allowed"))
	}

	dirPath := "./public/geojson"
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		return respondError(c, err)
	}

	baseName := strings.TrimSuffix(filepath.Base(fileHeader.Filename), filepath.Ext(fileHeader.Filename))
	baseName = invalidFileNameChars.ReplaceAllString(baseName, "_")
	baseName = strings.Trim(baseName, "_")
	if baseName == "" {
		baseName = "area_geojson"
	}

	storedName := fmt.Sprintf("%s_%d%s", baseName, time.Now().UnixNano(), ext)
	destination := filepath.Join(dirPath, storedName)

	src, err := fileHeader.Open()
	if err != nil {
		return respondError(c, err)
	}
	defer src.Close()

	dst, err := os.Create(destination)
	if err != nil {
		return respondError(c, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.Map{"geojson_file_path": "/geojson/" + storedName}, "GeoJSON uploaded successfully")
}

func (h *AreaHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")

	var req dto.UpdateAreaRequest
	if err := c.Bind().Body(&req); err != nil {
		return respondValidation(c, err)
	}

	if err := h.service.UpdateArea(id, req); err != nil {
		return respondError(c, err)
	}

	return respondOK(c, nil, "Area updated successfully")
}

func (h *AreaHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")

	if err := h.service.DeleteArea(id); err != nil {
		return respondError(c, err)
	}

	return respondOK(c, nil, "Area deleted successfully")
}
