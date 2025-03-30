package interfaces

import "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Bot interface {
	// Основные команды
	HandleCommand(update tgbotapi.Update)
	HandleCallbackQuery(update tgbotapi.Update)

	// Админ-панель
	ShowAdminPanel(chatID int64)
	IsAdmin(userID int64) bool
} 