package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"weveryone_bot_v2/bot"
	"weveryone_bot_v2/database"
	"weveryone_bot_v2/interfaces"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// saveUser сохраняет информацию о пользователе
func saveUser(db interfaces.Database, user *tgbotapi.User) error {
	if user == nil {
		return nil
	}
	return db.AddUser(user.ID, user.UserName)
}

// saveChat сохраняет информацию о чате
func saveChat(db interfaces.Database, chat *tgbotapi.Chat) error {
	if chat == nil {
		return nil
	}

	var title string
	switch chat.Type {
	case "private":
		// Для личных чатов используем имя пользователя
		title = fmt.Sprintf("👤 %s", chat.FirstName)
		if chat.LastName != "" {
			title += fmt.Sprintf(" %s", chat.LastName)
		}
	case "group":
		title = fmt.Sprintf("👥 %s", chat.Title)
	case "supergroup":
		title = fmt.Sprintf("👥 %s", chat.Title)
	case "channel":
		title = fmt.Sprintf("📢 %s", chat.Title)
	default:
		title = chat.Title
	}

	return db.AddChat(chat.ID, title)
}

func saveUserChatRelation(db interfaces.Database, user *tgbotapi.User, chat *tgbotapi.Chat) error {
	if user == nil || chat == nil {
		return nil
	}
	err := db.AddUserToChat(user.ID, chat.ID)
	// Игнорируем ошибку о том, что пользователь уже существует в чате
	if err != nil && strings.Contains(err.Error(), "пользователь уже существует в этом чате") {
		return nil
	}
	return err
}

func main() {
	// Инициализация базы данных
	db, err := database.NewSQLiteDB("data/bot.db")
	if err != nil {
		log.Fatal(err)
	}

	// Получение конфигурации
	botToken := getEnv("BOT_TOKEN", "5818425786:AAHU4OQYccUuhfRrJRg0UOjokjwIDDOa-jU")
	adminIDStr := getEnv("ADMIN_ID", "399040843")

	if botToken == "" || adminIDStr == "0" {
		log.Fatal("Переменные окружения BOT_TOKEN или ADMIN_ID не установлены")
	}

	adminID, err := strconv.ParseInt(adminIDStr, 10, 64)
	if err != nil || adminID == 0 {
		log.Fatal("Ошибка преобразования ADMIN_ID в число:", err)
	}

	// Создание бота
	telegramBot, err := bot.NewTelegramBot(botToken, adminID, db)
	if err != nil {
		log.Fatal(err)
	}

	// Настройка обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Обработка обновлений
	updates := telegramBot.GetUpdatesChan(u)
	log.Printf("start bot")
	for update := range updates {
		if update.InlineQuery != nil {
			telegramBot.HandleInlineQuery(update)
			continue
		}

		if update.Message != nil {
			// Сохраняем информацию о пользователе и чате
			if err := saveUser(db, update.Message.From); err != nil {
				log.Printf("Ошибка сохранения пользователя: %v", err)
			}
			if err := saveChat(db, update.Message.Chat); err != nil {
				log.Printf("Ошибка сохранения чата: %v", err)
			}
			// Сохраняем связь пользователя с чатом
			if err := saveUserChatRelation(db, update.Message.From, update.Message.Chat); err != nil {
				log.Printf("Ошибка сохранения связи пользователя с чатом: %v", err)
			}

			// Обрабатываем команду
			if update.Message.IsCommand() {
				telegramBot.HandleCommand(update)
			}
		}

		if update.CallbackQuery != nil {
			// Сохраняем информацию о пользователе и чате для callback query
			if err := saveUser(db, update.CallbackQuery.From); err != nil {
				log.Printf("Ошибка сохранения пользователя: %v", err)
			}
			if err := saveChat(db, update.CallbackQuery.Message.Chat); err != nil {
				log.Printf("Ошибка сохранения чата: %v", err)
			}
			// Сохраняем связь пользователя с чатом
			if err := saveUserChatRelation(db, update.CallbackQuery.From, update.CallbackQuery.Message.Chat); err != nil {
				log.Printf("Ошибка сохранения связи пользователя с чатом: %v", err)
			}

			// Обрабатываем callback query
			telegramBot.HandleCallbackQuery(update)
		}
	}
}
