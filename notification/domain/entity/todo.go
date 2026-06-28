package entity

import (
	"time"

	"gorm.io/gorm"
)

type Todo struct {
	ID          int            `gorm:"primaryKey;autoIncrement"`
	UserID      int
	Title       string
	Content     *string
	DueDate     *time.Time
	IsCompleted bool
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt
}
