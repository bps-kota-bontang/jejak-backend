package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type User struct {
	ID        string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Username  string         `gorm:"type:text;unique;not null"`
	Email     string         `gorm:"type:text;unique"`
	Password  *string        `gorm:"type:text"`
	Roles     pq.StringArray `gorm:"type:text[]"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
