package model

import (
	"time"
)

type User struct {
	ID           int64      `gorm:"primaryKey;autoIncrement"`
	Username     string     `gorm:"size:50;unique;not null"`
	PasswordHash string     `gorm:"size:255;not null;column:password_hash"`
	Role         string     `gorm:"size:20;default:'user'"`
	CreatedAt    *time.Time `gorm:"autoCreateTime;type:datetime"`
	UpdatedAt    *time.Time `gorm:"autoUpdateTime;type:datetime"`
	DeletedAt    *time.Time `gorm:"index;type:datetime"`
}
