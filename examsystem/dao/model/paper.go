package model

import (
	"time"
)

type Paper struct {
	ID          int64      `gorm:"primaryKey;autoIncrement"`
	Title       string     `gorm:"size:255;not null"`
	Description string     `gorm:"type:text;default:''"`
	TotalScore  int        `gorm:"default:100"`
	CreatorID   int64      `gorm:"not null;index"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `gorm:"index"`
}
