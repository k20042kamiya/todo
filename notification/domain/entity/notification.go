package entity

import "time"

const (
	NotificationTypeApproaching = "approaching"
	NotificationTypeOverdue     = "overdue"
)

type Notification struct {
	ID     int       `gorm:"primaryKey;autoIncrement"`
	TodoID int
	UserID int
	Type   string    `gorm:"size:32;not null"`
	SentAt time.Time `gorm:"autoCreateTime"`
}
