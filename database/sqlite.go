package database

import (
	"fmt"
	"weveryone_bot_v2/interfaces"
	"weveryone_bot_v2/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SQLiteDB struct {
	db *gorm.DB
}

// Проверяем, что SQLiteDB реализует интерфейс Database
var _ interfaces.Database = (*SQLiteDB)(nil)

func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия базы данных: %v", err)
	}

	// Автоматическая миграция схемы
	err = db.AutoMigrate(
		&models.User{},
		&models.Chat{},
		&models.Group{},
		&models.UserChat{},
		&models.UserGroup{},
		&models.GroupChat{},
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка миграции базы данных: %v", err)
	}

	return &SQLiteDB{db: db}, nil
}

// Реализация методов интерфейса Database
func (s *SQLiteDB) AddUser(userID int64, username string) error {
	if s.UserExists(userID) {
		return nil // Пользователь уже существует
	}
	user := models.User{
		UserID:   userID,
		Username: username,
	}
	return s.db.Create(&user).Error
}

func (s *SQLiteDB) DeleteUser(userID int64) error {
	return s.db.Where("user_id = ?", userID).Delete(&models.User{}).Error
}

func (s *SQLiteDB) ListUsers() []models.User {
	var users []models.User
	s.db.Find(&users)
	return users
}

func (s *SQLiteDB) AddChat(chatID int64, title string) error {
	if s.ChatExists(chatID) {
		return nil // Чат уже существует
	}
	chat := models.Chat{
		ChatID: chatID,
		Title:  title,
	}
	return s.db.Create(&chat).Error
}

func (s *SQLiteDB) DeleteChat(chatID int64) error {
	return s.db.Where("chat_id = ?", chatID).Delete(&models.Chat{}).Error
}

func (s *SQLiteDB) ListChats() []models.Chat {
	var chats []models.Chat
	s.db.Find(&chats)
	return chats
}

func (s *SQLiteDB) AddGroup(name string) error {
	if s.GroupExists(name) {
		return nil // Группа уже существует
	}
	group := models.Group{
		Name: name,
	}
	return s.db.Create(&group).Error
}

func (s *SQLiteDB) DeleteGroup(name string) error {
	return s.db.Where("name = ?", name).Delete(&models.Group{}).Error
}

func (s *SQLiteDB) ListGroups() []models.Group {
	var groups []models.Group
	s.db.Find(&groups)
	return groups
}

func (s *SQLiteDB) AddUserToChat(userID int64, chatID int64) error {
	// Проверяем существование пользователя и чата
	if !s.UserExists(userID) {
		return fmt.Errorf("пользователь не найден: %d", userID)
	}
	if !s.ChatExists(chatID) {
		return fmt.Errorf("чат не найден: %d", chatID)
	}

	// Проверяем существование связи
	var count int64
	s.db.Model(&models.UserChat{}).Where("user_id = ? AND chat_id = ?", userID, chatID).Count(&count)
	if count > 0 {
		return fmt.Errorf("пользователь уже существует в этом чате")
	}

	userChat := models.UserChat{
		UserID: userID,
		ChatID: chatID,
	}
	return s.db.Create(&userChat).Error
}

func (s *SQLiteDB) AddUserToGroup(userID int64, groupName string) error {
	// Проверяем существование пользователя и группы
	if !s.UserExists(userID) {
		return fmt.Errorf("пользователь не найден: %d", userID)
	}
	if !s.GroupExists(groupName) {
		return fmt.Errorf("группа не найдена: %s", groupName)
	}

	// Проверяем существование связи
	var count int64
	s.db.Model(&models.UserGroup{}).Where("user_id = ? AND group_name = ?", userID, groupName).Count(&count)
	if count > 0 {
		return nil // Связь уже существует
	}

	userGroup := models.UserGroup{
		UserID:    userID,
		GroupName: groupName,
	}
	return s.db.Create(&userGroup).Error
}

func (s *SQLiteDB) LinkGroupToChat(groupName string, chatID int64) error {
	// Проверяем существование группы и чата
	if !s.GroupExists(groupName) {
		return fmt.Errorf("группа не найдена: %s", groupName)
	}
	if !s.ChatExists(chatID) {
		return fmt.Errorf("чат не найден: %d", chatID)
	}

	// Проверяем существование связи
	var count int64
	s.db.Model(&models.GroupChat{}).Where("group_name = ? AND chat_id = ?", groupName, chatID).Count(&count)
	if count > 0 {
		return nil // Связь уже существует
	}

	groupChat := models.GroupChat{
		GroupName: groupName,
		ChatID:    chatID,
	}
	return s.db.Create(&groupChat).Error
}

func (s *SQLiteDB) GetUsersForMention(chatID int64, groupName string) []string {
	var users []models.User
	query := s.db.Joins("JOIN user_chats ON users.user_id = user_chats.user_id").
		Where("user_chats.chat_id = ?", chatID)

	if groupName != "" {
		query = query.Joins("JOIN user_groups ON users.user_id = user_groups.user_id").
			Where("user_groups.group_name = ?", groupName)
	}

	query.Find(&users)

	var usernames []string
	for _, user := range users {
		if user.Username != "" {
			usernames = append(usernames, "@"+user.Username)
		}
	}

	return usernames
}

// UserExists проверяет существование пользователя
func (s *SQLiteDB) UserExists(userID int64) bool {
	var count int64
	s.db.Model(&models.User{}).Where("user_id = ?", userID).Count(&count)
	return count > 0
}

// ChatExists проверяет существование чата
func (s *SQLiteDB) ChatExists(chatID int64) bool {
	var count int64
	s.db.Model(&models.Chat{}).Where("chat_id = ?", chatID).Count(&count)
	return count > 0
}

// GroupExists проверяет существование группы
func (s *SQLiteDB) GroupExists(name string) bool {
	var count int64
	s.db.Model(&models.Group{}).Where("name = ?", name).Count(&count)
	return count > 0
}

func (db *SQLiteDB) GetUser(userID int64) (*models.User, error) {
	var user models.User
	result := db.db.First(&user, "user_id = ?", userID)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (db *SQLiteDB) GetChat(chatID int64) (*models.Chat, error) {
	var chat models.Chat
	result := db.db.First(&chat, "chat_id = ?", chatID)
	if result.Error != nil {
		return nil, result.Error
	}
	return &chat, nil
}

func (db *SQLiteDB) GetGroup(name string) (*models.Group, error) {
	var group models.Group
	result := db.db.First(&group, "name = ?", name)
	if result.Error != nil {
		return nil, result.Error
	}
	return &group, nil
}

func (db *SQLiteDB) GetChatsForUser(userID int64) []models.Chat {
	var chats []models.Chat
	db.db.Joins("JOIN user_chats ON chats.chat_id = user_chats.chat_id").
		Where("user_chats.user_id = ?", userID).
		Find(&chats)
	return chats
}

func (db *SQLiteDB) GetGroupsForUser(userID int64) []models.Group {
	var groups []models.Group
	db.db.Joins("JOIN user_groups ON groups.name = user_groups.group_name").
		Where("user_groups.user_id = ?", userID).
		Find(&groups)
	return groups
}

func (db *SQLiteDB) GetGroupsForChat(chatID int64) []models.Group {
	var groups []models.Group
	db.db.Joins("JOIN group_chats ON groups.name = group_chats.group_name").
		Where("group_chats.chat_id = ?", chatID).
		Find(&groups)
	return groups
}

func (db *SQLiteDB) GetChatsForGroup(groupName string) []models.Chat {
	var chats []models.Chat
	db.db.Joins("JOIN group_chats ON chats.chat_id = group_chats.chat_id").
		Where("group_chats.group_name = ?", groupName).
		Find(&chats)
	return chats
}

func (db *SQLiteDB) GetUsersForChat(chatID int64) []models.User {
	var users []models.User
	db.db.Joins("JOIN user_chats ON users.user_id = user_chats.user_id").
		Where("user_chats.chat_id = ?", chatID).
		Find(&users)
	return users
}

func (db *SQLiteDB) AddUsersToChat(userIDs []int64, chatID int64) error {
	// Проверяем существование чата
	if !db.ChatExists(chatID) {
		return fmt.Errorf("чат не найден: %d", chatID)
	}

	// Проверяем существование всех пользователей
	for _, userID := range userIDs {
		if !db.UserExists(userID) {
			return fmt.Errorf("пользователь не найден: %d", userID)
		}
	}

	// Добавляем каждого пользователя в чат
	for _, userID := range userIDs {
		// Проверяем существование связи
		var count int64
		db.db.Model(&models.UserChat{}).Where("user_id = ? AND chat_id = ?", userID, chatID).Count(&count)
		if count == 0 {
			userChat := models.UserChat{
				UserID: userID,
				ChatID: chatID,
			}
			if err := db.db.Create(&userChat).Error; err != nil {
				return fmt.Errorf("ошибка добавления пользователя %d в чат: %v", userID, err)
			}
		}
	}

	return nil
}

func (s *SQLiteDB) AddUsersToGroup(userIDs []int64, groupName string) error {
	// Проверяем существование группы
	if !s.GroupExists(groupName) {
		return fmt.Errorf("группа не найдена: %s", groupName)
	}

	// Проверяем существование каждого пользователя и добавляем его в группу
	for _, userID := range userIDs {
		if !s.UserExists(userID) {
			return fmt.Errorf("пользователь не найден: %d", userID)
		}

		// Проверяем существование связи
		var count int64
		s.db.Model(&models.UserGroup{}).Where("user_id = ? AND group_name = ?", userID, groupName).Count(&count)
		if count > 0 {
			continue // Пропускаем, если связь уже существует
		}

		userGroup := models.UserGroup{
			UserID:    userID,
			GroupName: groupName,
		}
		if err := s.db.Create(&userGroup).Error; err != nil {
			return fmt.Errorf("ошибка добавления пользователя в группу: %v", err)
		}
	}

	return nil
}

func (db *SQLiteDB) GetUsersForGroup(groupName string) []models.User {
	var users []models.User
	db.db.Joins("JOIN user_groups ON users.user_id = user_groups.user_id").
		Where("user_groups.group_name = ?", groupName).
		Find(&users)
	return users
}
