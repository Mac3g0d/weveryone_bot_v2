package models

import "gorm.io/gorm"

type Group struct {
	gorm.Model
	Name     string `gorm:"uniqueIndex"`
	Users    []User `gorm:"many2many:group_users;"`
	Chats    []Chat `gorm:"many2many:chat_groups;"`
} 