package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	UserID   int64  `gorm:"uniqueIndex"`
	Username string
} 