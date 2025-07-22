package model

import (
	"time"
)

type PaperQuestion struct {
	ID            int64      `gorm:"primaryKey;autoIncrement"`
	PaperID       int64      `gorm:"not null;index"`
	QuestionID    int64      `gorm:"not null;index"`
	QuestionOrder int        `gorm:"not null"`
	Score         int        `gorm:"default:5"`
	CreatedAt     time.Time  `gorm:"autoCreateTime"`
	DeletedAt     *time.Time `gorm:"index"`
}
