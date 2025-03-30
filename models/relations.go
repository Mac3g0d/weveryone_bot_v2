package models

import "time"

// UserChat представляет связь между пользователем и чатом
type UserChat struct {
	UserID    int64     `gorm:"primaryKey"`
	ChatID    int64     `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

// UserGroup представляет связь между пользователем и группой
type UserGroup struct {
	UserID    int64     `gorm:"primaryKey"`
	GroupName string    `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

// GroupChat представляет связь между группой и чатом
type GroupChat struct {
	GroupName string    `gorm:"primaryKey"`
	ChatID    int64     `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
} 