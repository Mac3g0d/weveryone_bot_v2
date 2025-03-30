package models

import "gorm.io/gorm"

type Chat struct {
	gorm.Model
	ChatID   int64  `gorm:"uniqueIndex"`
	Title    string
	Users    []User `gorm:"many2many:chat_users;"`
} 