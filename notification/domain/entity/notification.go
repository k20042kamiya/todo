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
	Type   string
	SentAt time.Time `gorm:"autoCreateTime"`
}
