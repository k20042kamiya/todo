package todo

import (
	"time"

	"gorm.io/gorm"
)

type Todo struct {
	ID          int            `gorm:"primaryKey;autoIncrement"`
	UserID      int            `gorm:"not null;index"`
	Title       string         `gorm:"size:255;not null"`
	Content     *string        `gorm:"type:text"`
	DueDate     *time.Time     `gorm:"type:date"`
	IsCompleted bool           `gorm:"default:false"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
