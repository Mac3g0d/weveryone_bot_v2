package interfaces

import "weveryone_bot_v2/models"

// Database определяет интерфейс для работы с базой данных
type Database interface {
	// Методы для работы с пользователями
	AddUser(userID int64, username string) error
	DeleteUser(userID int64) error
	ListUsers() []models.User
	UserExists(userID int64) bool
	GetUser(userID int64) (*models.User, error)
	GetChatsForUser(userID int64) []models.Chat
	GetGroupsForUser(userID int64) []models.Group

	// Методы для работы с чатами
	AddChat(chatID int64, title string) error
	DeleteChat(chatID int64) error
	ListChats() []models.Chat
	ChatExists(chatID int64) bool
	GetChat(chatID int64) (*models.Chat, error)
	GetUsersForMention(chatID int64, groupName string) []string
	GetGroupsForChat(chatID int64) []models.Group

	// Методы для работы с группами
	AddGroup(name string) error
	DeleteGroup(name string) error
	ListGroups() []models.Group
	GroupExists(name string) bool
	GetGroup(name string) (*models.Group, error)
	GetChatsForGroup(groupName string) []models.Chat
	GetUsersForGroup(groupName string) []models.User

	// Методы для работы со связями
	AddUserToChat(userID int64, chatID int64) error
	AddUserToGroup(userID int64, groupName string) error
	LinkGroupToChat(groupName string, chatID int64) error
	GetUsersForChat(chatID int64) []models.User
	AddUsersToChat(userIDs []int64, chatID int64) error
	AddUsersToGroup(userIDs []int64, groupName string) error
} 