package dto

import "time"

type CreateAreaRequest struct {
	Name            string   `json:"name" validate:"required"`
	GeoJSONFilePath string   `json:"geojson_file_path" validate:"required"`
	ListKeys        []string `json:"list_keys" validate:"required,min=1,dive,required"`
	Description     *string  `json:"description,omitempty"`
}

type UpdateAreaRequest struct {
	Name            string   `json:"name,omitempty"`
	GeoJSONFilePath string   `json:"geojson_file_path,omitempty"`
	ListKeys        []string `json:"list_keys,omitempty"`
	Description     *string  `json:"description,omitempty"`
}

type AreaResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	GeoJSONFilePath string    `json:"geojson_file_path"`
	ListKeys        []string  `json:"list_keys"`
	Description     *string   `json:"description,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
