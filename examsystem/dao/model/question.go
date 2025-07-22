package model

import (
	"time"

	"gorm.io/gorm"
)

type QuestionType string

const (
	QuestionTypeSingle   QuestionType = "single"
	QuestionTypeMultiple QuestionType = "multiple"
)

type Question struct {
	ID           int64          `gorm:"primaryKey;autoIncrement"`
	Title        string         `gorm:"type:text;not null"`
	QuestionType QuestionType   `gorm:"size:20;not null;check:question_type IN ('single','multiple')"`
	Options      string         `gorm:"type:text;not null"`
	Answer       string         `gorm:"type:text;not null"`
	Explanation  string         `gorm:"type:text;default:''"`
	Keywords     string         `gorm:"size:255;default:''"`
	Language     string         `gorm:"size:50;not null"`
	AIModel      string         `gorm:"size:50;not null;column:ai_model"`
	UserID       int64          `gorm:"not null;index"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index;"`
}
