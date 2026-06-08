package models

import (
	"time"

	"gorm.io/gorm"
)

type Area struct {
	ID              string   `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name            string   `gorm:"type:text;not null"`
	GeoJSONFilePath string   `gorm:"type:text;not null"` // Path to uploaded geojson file
	ListKeys        []string `gorm:"type:json;serializer:json;not null;default:'[]'"`
	Description     *string  `gorm:"type:text"`
	Surveys         []Survey `gorm:"foreignKey:AreaID;references:ID"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}
