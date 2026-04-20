package user

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          int            `gorm:"primaryKey;autoIncrement"`
	FirebaseUID string         `gorm:"column:firebase_uid;size:128;not null;uniqueIndex"`
	Email       string         `gorm:"size:255;not null;uniqueIndex"`
	Name        string         `gorm:"size:100;not null"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
