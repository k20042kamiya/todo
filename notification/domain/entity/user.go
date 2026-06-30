package entity

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          int            `gorm:"primaryKey;autoIncrement"`
	FirebaseUID string         `gorm:"column:firebase_uid"`
	Email       string
	Name        string
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt
}
