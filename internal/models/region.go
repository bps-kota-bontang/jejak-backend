package models

import (
	"time"

	"gorm.io/gorm"
)

type Region struct {
	ID              string  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SurveyID        string  `gorm:"type:text;not null;index;uniqueIndex:idx_region_unique"`
	SurveyPeriodID  string  `gorm:"type:text;not null;index;uniqueIndex:idx_region_unique"`
	RegionGroupID   string  `gorm:"type:text;not null;index"`
	AssignmentCount int     `gorm:"type:int;not null;default:0"`
	Usaha           int     `gorm:"type:int;not null;default:0"`
	OpenCount       int     `gorm:"type:int;not null;default:0"`
	DraftCount      int     `gorm:"type:int;not null;default:0"`
	SubmittedCount  int     `gorm:"type:int;not null;default:0"`
	ApprovedCount   int     `gorm:"type:int;not null;default:0"`
	RejectedCount   int     `gorm:"type:int;not null;default:0"`
	RevokedCount    int     `gorm:"type:int;not null;default:0"`
	Level1          *string `gorm:"type:text;index"`
	Level1Label     *string `gorm:"type:text"`
	Level2          *string `gorm:"type:text;index"`
	Level2Label     *string `gorm:"type:text"`
	Level3          *string `gorm:"type:text;index"`
	Level3Label     *string `gorm:"type:text"`
	Level4          *string `gorm:"type:text;index"`
	Level4Label     *string `gorm:"type:text"`
	Level5          *string `gorm:"type:text;index"`
	Level5Label     *string `gorm:"type:text"`
	Level6          *string `gorm:"type:text;index"`
	Level6Label     *string `gorm:"type:text"`
	PJ              *string `gorm:"type:text"`
	PML             *string `gorm:"type:text"`
	PPL             *string `gorm:"type:text"`
	FullCode        string  `gorm:"type:text;not null;index;uniqueIndex:idx_region_unique"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}
