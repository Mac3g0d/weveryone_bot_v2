package bot

import (
	"fmt"
	"strconv"
	"strings"
	"weveryone_bot_v2/interfaces"
	"weveryone_bot_v2/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	itemsPerPage = 15
)

type TelegramBot struct {
	bot     *tgbotapi.BotAPI
	db      interfaces.Database
	adminID int64
}

func NewTelegramBot(token string, adminID int64, db interfaces.Database) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания бота: %v", err)
	}

	return &TelegramBot{
		bot:     bot,
		db:      db,
		adminID: adminID,
	}, nil
}

// GetUpdatesChan возвращает канал обновлений от Telegram
func (b *TelegramBot) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	return b.bot.GetUpdatesChan(config)
}

func (b *TelegramBot) IsAdmin(userID int64) bool {
	return userID == b.adminID
}

func (b *TelegramBot) ShowAdminPanel(chatID int64) {
	helpText := `Доступные команды:

Основные команды:
/all или /everyone - упомянуть всех пользователей в чате
/group <название> - упомянуть пользователей определенной группы
/help - показать это сообщение
/start - показать админ-панель (только для администраторов)

Команды администратора:
/admin - показать админ-панель
/add_user <user_id> <username> - добавить пользователя
/del_user <user_id> - удалить пользователя
/list_users - показать список пользователей
/add_chat <chat_id> <title> - добавить чат
/del_chat <chat_id> - удалить чат
/list_chats - показать список чатов
/add_group <name> - создать группу
/del_group <name> - удалить группу
/list_groups - показать список групп
/add_to_chat <user_id> <chat_id> - добавить пользователя в чат
/add_to_group <user_id> <group_name> - добавить пользователя в группу
/link_group_chat <group_name> <chat_id> - связать группу с чатом
/add_users_to_chat <chat_id> <user_id1> [user_id2 ...] - добавить несколько пользователей в чат`

	msg := tgbotapi.NewMessage(chatID, helpText)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👥 Пользователи", "admin_users"),
			tgbotapi.NewInlineKeyboardButtonData("➕ Создать пользователя", "create_user"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💬 Чаты", "admin_chats"),
			tgbotapi.NewInlineKeyboardButtonData("➕ Создать чат", "create_chat"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👥 Группы", "admin_groups"),
			tgbotapi.NewInlineKeyboardButtonData("➕ Создать группу", "create_group"),
		),
	)
	b.bot.Send(msg)
}

func (b *TelegramBot) ShowHelp(chatID int64) {
	helpText := `Доступные команды:

Основные команды:
/all или /everyone - упомянуть всех пользователей в чате
/group <название> - упомянуть пользователей определенной группы
/help - показать это сообщение
/start - показать админ-панель (только для администраторов)

Команды администратора:
/admin - показать админ-панель
/add_user <user_id> <username> - добавить пользователя
/del_user <user_id> - удалить пользователя
/list_users - показать список пользователей
/add_chat <chat_id> <title> - добавить чат
/del_chat <chat_id> - удалить чат
/list_chats - показать список чатов
/add_group <name> - создать группу
/del_group <name> - удалить группу
/list_groups - показать список групп
/add_to_chat <user_id> <chat_id> - добавить пользователя в чат
/add_to_group <user_id> <group_name> - добавить пользователя в группу
/link_group_chat <group_name> <chat_id> - связать группу с чатом`

	msg := tgbotapi.NewMessage(chatID, helpText)
	b.bot.Send(msg)
}

func (b *TelegramBot) ShowViewMenu(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Выберите, что хотите просмотреть:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Просмотр пользователей", "view_users"),
			tgbotapi.NewInlineKeyboardButtonData("Просмотр чатов", "view_chats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Просмотр групп", "view_groups"),
			tgbotapi.NewInlineKeyboardButtonData("Просмотр связей", "view_relations"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", "admin_back"),
		),
	)
	b.bot.Send(msg)
}

// deleteMessage удаляет предыдущее сообщение
func (b *TelegramBot) deleteMessage(chatID int64, messageID int) {
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	b.bot.Send(deleteMsg)
}

func (b *TelegramBot) ShowUsersList(chatID int64, page int, update *tgbotapi.Update) {
	// Удаляем предыдущее сообщение
	if update != nil && update.CallbackQuery != nil {
		b.deleteMessage(chatID, update.CallbackQuery.Message.MessageID)
	}

	users := b.db.ListUsers()
	totalPages := (len(users) + itemsPerPage - 1) / itemsPerPage
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * itemsPerPage
	end := start + itemsPerPage
	if end > len(users) {
		end = len(users)
	}

	var msgText strings.Builder
	msgText.WriteString(fmt.Sprintf("Список пользователей (страница %d из %d):\n\n", page, totalPages))
	for _, user := range users[start:end] {
		msgText.WriteString(fmt.Sprintf("ID: %d\nUsername: %s\n\n", user.UserID, user.Username))
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	row := make([]tgbotapi.InlineKeyboardButton, 0)

	if page > 1 {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("◀️", fmt.Sprintf("users_page_%d", page-1)))
	}
	if page < totalPages {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("▶️", fmt.Sprintf("users_page_%d", page+1)))
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}

	for _, user := range users[start:end] {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("👤 %s", user.Username),
				fmt.Sprintf("user_info_%d", user.UserID),
			),
		})
	}

	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("➕ Создать пользователя", "create_user"),
		tgbotapi.NewInlineKeyboardButtonData("Назад", "admin_back"),
	})

	msg := tgbotapi.NewMessage(chatID, msgText.String())
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	b.bot.Send(msg)
}

func (b *TelegramBot) ShowChatsList(chatID int64, page int, update *tgbotapi.Update) {
	// Удаляем предыдущее сообщение
	if update != nil && update.CallbackQuery != nil {
		b.deleteMessage(chatID, update.CallbackQuery.Message.MessageID)
	}

	chats := b.db.ListChats()
	totalPages := (len(chats) + itemsPerPage - 1) / itemsPerPage
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * itemsPerPage
	end := start + itemsPerPage
	if end > len(chats) {
		end = len(chats)
	}

	var msgText strings.Builder
	msgText.WriteString(fmt.Sprintf("Список чатов (страница %d из %d):\n\n", page, totalPages))
	for _, chat := range chats[start:end] {
		msgText.WriteString(fmt.Sprintf("ID: %d\nTitle: %s\n\n", chat.ChatID, chat.Title))
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	row := make([]tgbotapi.InlineKeyboardButton, 0)

	if page > 1 {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("◀️", fmt.Sprintf("chats_page_%d", page-1)))
	}
	if page < totalPages {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("▶️", fmt.Sprintf("chats_page_%d", page+1)))
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}

	for _, chat := range chats[start:end] {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("💬 %s", chat.Title),
				fmt.Sprintf("chat_info_%d", chat.ChatID),
			),
		})
	}

	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("➕ Создать чат", "create_chat"),
		tgbotapi.NewInlineKeyboardButtonData("Назад", "admin_back"),
	})

	msg := tgbotapi.NewMessage(chatID, msgText.String())
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	b.bot.Send(msg)
}

func (b *TelegramBot) ShowGroupsList(chatID int64, page int, update *tgbotapi.Update) {
	// Удаляем предыдущее сообщение
	if update != nil && update.CallbackQuery != nil {
		b.deleteMessage(chatID, update.CallbackQuery.Message.MessageID)
	}

	groups := b.db.ListGroups()
	totalPages := (len(groups) + itemsPerPage - 1) / itemsPerPage
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * itemsPerPage
	end := start + itemsPerPage
	if end > len(groups) {
		end = len(groups)
	}

	var msgText strings.Builder
	msgText.WriteString(fmt.Sprintf("Список групп (страница %d из %d):\n\n", page, totalPages))
	for _, group := range groups[start:end] {
		msgText.WriteString(fmt.Sprintf("Name: %s\n\n", group.Name))
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	row := make([]tgbotapi.InlineKeyboardButton, 0)

	if page > 1 {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("◀️", fmt.Sprintf("groups_page_%d", page-1)))
	}
	if page < totalPages {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("▶️", fmt.Sprintf("groups_page_%d", page+1)))
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}

	for _, group := range groups[start:end] {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("👥 %s", group.Name),
				fmt.Sprintf("group_info_%s", group.Name),
			),
		})
	}

	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("➕ Создать группу", "create_group"),
		tgbotapi.NewInlineKeyboardButtonData("Назад", "admin_back"),
	})

	msg := tgbotapi.NewMessage(chatID, msgText.String())
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	b.bot.Send(msg)
}

func (b *TelegramBot) ShowUserInfo(chatID int64, userID int64) {
	user, err := b.db.GetUser(userID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Ошибка получения информации о пользователе")
		b.bot.Send(msg)
		return
	}

	var msgText strings.Builder
	msgText.WriteString(fmt.Sprintf("Информация о пользователе:\n\n"))
	msgText.WriteString(fmt.Sprintf("ID: %d\n", user.UserID))
	msgText.WriteString(fmt.Sprintf("Username: %s\n", user.Username))

	// Получаем чаты пользователя
	chats := b.db.GetChatsForUser(userID)
	if len(chats) > 0 {
		msgText.WriteString("\nЧаты пользователя:\n")
		for _, chat := range chats {
			msgText.WriteString(fmt.Sprintf("- %s\n", chat.Title))
		}
	}

	// Получаем группы пользователя
	groups := b.db.GetGroupsForUser(userID)
	if len(groups) > 0 {
		msgText.WriteString("\nГруппы пользователя:\n")
		for _, group := range groups {
			msgText.WriteString(fmt.Sprintf("- %s\n", group.Name))
		}
	}

	rows := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("✏️ Редактировать", fmt.Sprintf("edit_user_%d", userID)),
			tgbotapi.NewInlineKeyboardButtonData("🗑 Удалить", fmt.Sprintf("delete_user_%d", userID)),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("Назад", "admin_users"),
		},
	}

	msg := tgbotapi.NewMessage(chatID, msgText.String())
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	b.bot.Send(msg)
}

func (b *TelegramBot) ShowChatInfo(chatID int64, targetChatID int64) {
	chat, err := b.db.GetChat(targetChatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Ошибка получения информации о чате")
		b.bot.Send(msg)
		return
	}

	// Получаем список пользователей чата
	users := b.db.GetUsersForChat(targetChatID)
	var usersList string
	if len(users) > 0 {
		usersList = "\n\nПользователи чата:\n"
		for _, user := range users {
			usersList += fmt.Sprintf("- @%s\n", user.Username)
		}
	} else {
		usersList = "\n\nВ чате пока нет пользователей"
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Информация о чате:\nID: %d\nНазвание: %s%s",
		chat.ChatID, chat.Title, usersList))

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Добавить пользователей", fmt.Sprintf("add_users_to_chat_%d", targetChatID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ Редактировать", fmt.Sprintf("edit_chat_%d", targetChatID)),
			tgbotapi.NewInlineKeyboardButtonData("🗑 Удалить", fmt.Sprintf("delete_chat_%d", targetChatID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", "admin_chats"),
		),
	)
	b.bot.Send(msg)
}

func (b *TelegramBot) ShowUsersToAddToChat(chatID int64, targetChatID int64, page int, update *tgbotapi.Update) {
	// Удаляем предыдущее сообщение
	if update != nil && update.CallbackQuery != nil {
		b.deleteMessage(chatID, update.CallbackQuery.Message.MessageID)
	}

	// Получаем всех пользователей
	allUsers := b.db.ListUsers()

	// Получаем пользователей, которые уже в чате
	chatUsers := b.db.GetUsersForChat(targetChatID)
	chatUserMap := make(map[int64]bool)
	for _, user := range chatUsers {
		chatUserMap[user.UserID] = true
	}

	// Фильтруем пользователей, которых еще нет в чате
	var availableUsers []models.User
	for _, user := range allUsers {
		if !chatUserMap[user.UserID] {
			availableUsers = append(availableUsers, user)
		}
	}

	// Если нет доступных пользователей, показываем сообщение и возвращаемся
	if len(availableUsers) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Нет доступных пользователей для добавления в чат")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("chat_info_%d", targetChatID)),
			),
		)
		b.bot.Send(msg)
		return
	}

	totalPages := (len(availableUsers) + itemsPerPage - 1) / itemsPerPage
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * itemsPerPage
	end := start + itemsPerPage
	if end > len(availableUsers) {
		end = len(availableUsers)
	}

	var msgText strings.Builder
	msgText.WriteString(fmt.Sprintf("Выберите пользователей для добавления в чат (страница %d из %d):\n\n", page, totalPages))
	for _, user := range availableUsers[start:end] {
		msgText.WriteString(fmt.Sprintf("ID: %d\nUsername: %s\n\n", user.UserID, user.Username))
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	row := make([]tgbotapi.InlineKeyboardButton, 0)

	if page > 1 {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("◀️", fmt.Sprintf("add_users_to_chat_page_%d_%d", targetChatID, page-1)))
	}
	if page < totalPages {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("▶️", fmt.Sprintf("add_users_to_chat_page_%d_%d", targetChatID, page+1)))
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}

	for _, user := range availableUsers[start:end] {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("➕ @%s", user.Username),
				fmt.Sprintf("add_user_to_chat_%d_%d", targetChatID, user.UserID),
			),
		})
	}

	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("chat_info_%d", targetChatID)),
	})

	msg := tgbotapi.NewMessage(chatID, msgText.String())
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	b.bot.Send(msg)
}

func (b *TelegramBot) ShowGroupInfo(chatID int64, groupName string) {
	group, err := b.db.GetGroup(groupName)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Ошибка получения информации о группе: %v", err))
		b.bot.Send(msg)
		return
	}

	users := b.db.GetUsersForGroup(groupName)
	var msgText strings.Builder
	msgText.WriteString(fmt.Sprintf("Информация о группе: %s\n\n", group.Name))
	msgText.WriteString("Пользователи в группе:\n")
	for _, user := range users {
		msgText.WriteString(fmt.Sprintf("- @%s\n", user.Username))
	}

	msg := tgbotapi.NewMessage(chatID, msgText.String())
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Добавить пользователей", fmt.Sprintf("add_users_to_group_%s", groupName)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ Редактировать", fmt.Sprintf("edit_group_%s", groupName)),
			tgbotapi.NewInlineKeyboardButtonData("🗑 Удалить", fmt.Sprintf("delete_group_%s", groupName)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", "admin_groups"),
		),
	)
	b.bot.Send(msg)
}

func (b *TelegramBot) ShowUsersToAddToGroup(chatID int64, groupName string, page int, update *tgbotapi.Update) {
	// Удаляем предыдущее сообщение
	if update != nil && update.CallbackQuery != nil {
		b.deleteMessage(chatID, update.CallbackQuery.Message.MessageID)
	}

	// Получаем всех пользователей
	allUsers := b.db.ListUsers()

	// Получаем пользователей, которые уже в группе
	groupUsers := b.db.GetUsersForGroup(groupName)
	groupUserMap := make(map[int64]bool)
	for _, user := range groupUsers {
		groupUserMap[user.UserID] = true
	}

	// Фильтруем пользователей, которых еще нет в группе
	var availableUsers []models.User
	for _, user := range allUsers {
		if !groupUserMap[user.UserID] {
			availableUsers = append(availableUsers, user)
		}
	}

	// Если нет доступных пользователей, показываем сообщение и возвращаемся
	if len(availableUsers) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Нет доступных пользователей для добавления в группу")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("group_info_%s", groupName)),
			),
		)
		b.bot.Send(msg)
		return
	}

	totalPages := (len(availableUsers) + itemsPerPage - 1) / itemsPerPage
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * itemsPerPage
	end := start + itemsPerPage
	if end > len(availableUsers) {
		end = len(availableUsers)
	}

	var msgText strings.Builder
	msgText.WriteString(fmt.Sprintf("Выберите пользователей для добавления в группу %s (страница %d из %d):\n\n", groupName, page, totalPages))
	for _, user := range availableUsers[start:end] {
		msgText.WriteString(fmt.Sprintf("ID: %d\nUsername: %s\n\n", user.UserID, user.Username))
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	row := make([]tgbotapi.InlineKeyboardButton, 0)

	if page > 1 {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("◀️", fmt.Sprintf("add_users_to_group_page_%s_%d", groupName, page-1)))
	}
	if page < totalPages {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("▶️", fmt.Sprintf("add_users_to_group_page_%s_%d", groupName, page+1)))
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}

	// Добавляем кнопки для каждого пользователя
	for _, user := range availableUsers[start:end] {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("➕ Добавить @%s", user.Username),
				fmt.Sprintf("add_user_to_group_%s_%d", groupName, user.UserID)),
		))
	}

	// Добавляем кнопку "Назад"
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("group_info_%s", groupName)),
	))

	msg := tgbotapi.NewMessage(chatID, msgText.String())
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	b.bot.Send(msg)
}

func (b *TelegramBot) ShowRelations(chatID int64) {
	chats := b.db.ListChats()
	groups := b.db.ListGroups()

	var msgText strings.Builder
	msgText.WriteString("Связи в системе:\n\n")

	msgText.WriteString("Пользователи в чатах:\n")
	for _, chat := range chats {
		users := b.db.GetUsersForMention(chat.ChatID, "")
		if len(users) > 0 {
			msgText.WriteString(fmt.Sprintf("Чат %s (%d): %s\n", chat.Title, chat.ChatID, strings.Join(users, ", ")))
		}
	}
	msgText.WriteString("\n")

	msgText.WriteString("Пользователи в группах:\n")
	for _, group := range groups {
		users := b.db.GetUsersForMention(0, group.Name)
		if len(users) > 0 {
			msgText.WriteString(fmt.Sprintf("Группа %s: %s\n", group.Name, strings.Join(users, ", ")))
		}
	}

	msg := tgbotapi.NewMessage(chatID, msgText.String())
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", "admin_view"),
		),
	)
	b.bot.Send(msg)
}

func (b *TelegramBot) HandleCommand(update tgbotapi.Update) {
	msg := update.Message
	chatID := msg.Chat.ID
	userID := msg.From.ID
	command := msg.Command()

	// Добавляем инлайн-подсказки для команд
	if command == "" {
		text := msg.Text
		if strings.HasPrefix(text, "/") {
			// Показываем подсказки для команд
			commands := []string{
				"/all - упомянуть всех пользователей в чате",
				"/everyone - упомянуть всех пользователей в чате",
				"/group - упомянуть пользователей определенной группы",
				"/help - показать справку",
				"/start - показать админ-панель",
				"/admin - показать админ-панель",
				"/add_user - добавить пользователя",
				"/del_user - удалить пользователя",
				"/list_users - показать список пользователей",
				"/add_chat - добавить чат",
				"/del_chat - удалить чат",
				"/list_chats - показать список чатов",
				"/add_group - создать группу",
				"/del_group - удалить группу",
				"/list_groups - показать список групп",
				"/add_to_chat - добавить пользователя в чат",
				"/add_to_group - добавить пользователя в группу",
				"/link_group_chat - связать группу с чатом",
				"/add_users_to_chat - добавить несколько пользователей в чат",
			}

			var suggestions []string
			partial := strings.ToLower(text[1:]) // Убираем "/" и приводим к нижнему регистру

			// Определяем тип команды и показываем соответствующие подсказки
			switch {
			case strings.HasPrefix(partial, "add_to_chat"):
				chats := b.db.ListChats()
				for _, chat := range chats {
					suggestions = append(suggestions, fmt.Sprintf("/add_to_chat %d", chat.ChatID))
				}
			case strings.HasPrefix(partial, "add_to_group"):
				groups := b.db.ListGroups()
				for _, group := range groups {
					suggestions = append(suggestions, fmt.Sprintf("/add_to_group %s", group.Name))
				}
			case strings.HasPrefix(partial, "del_user"):
				users := b.db.ListUsers()
				for _, user := range users {
					suggestions = append(suggestions, fmt.Sprintf("/del_user %d", user.UserID))
				}
			case strings.HasPrefix(partial, "del_chat"):
				chats := b.db.ListChats()
				for _, chat := range chats {
					suggestions = append(suggestions, fmt.Sprintf("/del_chat %d", chat.ChatID))
				}
			case strings.HasPrefix(partial, "del_group"):
				groups := b.db.ListGroups()
				for _, group := range groups {
					suggestions = append(suggestions, fmt.Sprintf("/del_group %s", group.Name))
				}
			case strings.HasPrefix(partial, "link_group_chat"):
				groups := b.db.ListGroups()
				chats := b.db.ListChats()
				for _, group := range groups {
					for _, chat := range chats {
						suggestions = append(suggestions, fmt.Sprintf("/link_group_chat %s %d", group.Name, chat.ChatID))
					}
				}
			case strings.HasPrefix(partial, "add_users_to_chat"):
				chats := b.db.ListChats()
				for _, chat := range chats {
					suggestions = append(suggestions, fmt.Sprintf("/add_users_to_chat %d", chat.ChatID))
				}
			default:
				for _, cmd := range commands {
					if strings.HasPrefix(strings.ToLower(cmd), partial) {
						suggestions = append(suggestions, cmd)
					}
				}
			}

			if len(suggestions) > 0 {
				msg := tgbotapi.NewMessage(chatID, "Возможно, вы имели в виду:\n\n"+strings.Join(suggestions, "\n"))
				b.bot.Send(msg)
				return
			}
		}
	}

	switch command {
	case "start":
		if b.IsAdmin(userID) {
			b.ShowAdminPanel(chatID)
		} else {
			msg := tgbotapi.NewMessage(chatID, "У вас нет доступа к этой функции.")
			b.bot.Send(msg)
		}

	case "help":
		b.ShowHelp(chatID)

	case "all", "everyone":
		users := b.db.GetUsersForMention(chatID, "")
		if len(users) > 0 {
			msg := tgbotapi.NewMessage(chatID, strings.Join(users, " "))
			b.bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(chatID, "В этом чате пока нет пользователей.")
			b.bot.Send(msg)
		}

	case "group":
		args := strings.Fields(msg.Text)
		if len(args) < 2 {
			msg := tgbotapi.NewMessage(chatID, "Использование: /group <название_группы>")
			b.bot.Send(msg)
			return
		}
		groupName := args[1]
		users := b.db.GetUsersForMention(chatID, groupName)
		if len(users) > 0 {
			msg := tgbotapi.NewMessage(chatID, strings.Join(users, " "))
			b.bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(chatID, "В этой группе пока нет пользователей.")
			b.bot.Send(msg)
		}

	case "add_user":
		if !b.IsAdmin(userID) {
			msg := tgbotapi.NewMessage(chatID, "У вас нет доступа к этой функции.")
			b.bot.Send(msg)
			return
		}
		args := strings.Fields(msg.Text)
		if len(args) != 3 {
			msg := tgbotapi.NewMessage(chatID, "Использование: /add_user <user_id> <username>")
			b.bot.Send(msg)
			return
		}
		userID, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "Неверный формат user_id")
			b.bot.Send(msg)
			return
		}
		if err := b.db.AddUser(userID, args[2]); err != nil {
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Ошибка добавления пользователя: %v", err))
			b.bot.Send(msg)
			return
		}
		msg := tgbotapi.NewMessage(chatID, "Пользователь успешно добавлен")
		b.bot.Send(msg)

	case "add_chat":
		if !b.IsAdmin(userID) {
			msg := tgbotapi.NewMessage(chatID, "У вас нет доступа к этой функции.")
			b.bot.Send(msg)
			return
		}
		args := strings.Fields(msg.Text)
		if len(args) != 3 {
			msg := tgbotapi.NewMessage(chatID, "Использование: /add_chat <chat_id> <title>")
			b.bot.Send(msg)
			return
		}
		chatID, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "Неверный формат chat_id")
			b.bot.Send(msg)
			return
		}
		if err := b.db.AddChat(chatID, args[2]); err != nil {
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Ошибка добавления чата: %v", err))
			b.bot.Send(msg)
			return
		}
		msg := tgbotapi.NewMessage(chatID, "Чат успешно добавлен")
		b.bot.Send(msg)

	case "add_group":
		if !b.IsAdmin(userID) {
			msg := tgbotapi.NewMessage(chatID, "У вас нет доступа к этой функции.")
			b.bot.Send(msg)
			return
		}
		args := strings.Fields(msg.Text)
		if len(args) != 2 {
			msg := tgbotapi.NewMessage(chatID, "Использование: /add_group <name>")
			b.bot.Send(msg)
			return
		}
		if err := b.db.AddGroup(args[1]); err != nil {
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Ошибка добавления группы: %v", err))
			b.bot.Send(msg)
			return
		}
		msg := tgbotapi.NewMessage(chatID, "Группа успешно добавлена")
		b.bot.Send(msg)
		b.ShowAdminPanel(chatID)

	case "add_users_to_chat":
		if !b.IsAdmin(userID) {
			msg := tgbotapi.NewMessage(chatID, "У вас нет доступа к этой функции.")
			b.bot.Send(msg)
			return
		}
		args := strings.Fields(msg.Text)
		if len(args) < 3 {
			msg := tgbotapi.NewMessage(chatID, "Использование: /add_users_to_chat <chat_id> <user_id1> [user_id2 ...]")
			b.bot.Send(msg)
			return
		}
		chatID, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "Неверный формат chat_id")
			b.bot.Send(msg)
			return
		}
		var userIDs []int64
		for _, arg := range args[2:] {
			userID, err := strconv.ParseInt(arg, 10, 64)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Неверный формат user_id: %s", arg))
				b.bot.Send(msg)
				return
			}
			userIDs = append(userIDs, userID)
		}
		if err := b.db.AddUsersToChat(userIDs, chatID); err != nil {
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Ошибка добавления пользователей в чат: %v", err))
			b.bot.Send(msg)
			return
		}
		msg := tgbotapi.NewMessage(chatID, "Пользователи успешно добавлены в чат")
		b.bot.Send(msg)

	default:
		msg := tgbotapi.NewMessage(chatID, "Неизвестная команда. Используйте /help для просмотра доступных команд.")
		b.bot.Send(msg)
	}
}

func (b *TelegramBot) HandleCallbackQuery(update tgbotapi.Update) {
	query := update.CallbackQuery.Data
	adminchatID := update.CallbackQuery.Message.Chat.ID
	userID := update.CallbackQuery.From.ID

	if !b.IsAdmin(userID) {
		msg := tgbotapi.NewMessage(adminchatID, "У вас нет доступа к этой функции.")
		b.bot.Send(msg)
		return
	}

	// Удаляем предыдущее сообщение
	b.deleteMessage(adminchatID, update.CallbackQuery.Message.MessageID)

	switch {
	case query == "all":
		// Получаем все чаты пользователя
		chats := b.db.GetChatsForUser(userID)
		if len(chats) > 0 {
			// Используем первый чат из списка
			users := b.db.GetUsersForMention(chats[0].ChatID, "")
			if len(users) > 0 {
				msg := tgbotapi.NewMessage(adminchatID, strings.Join(users, " "))
				b.bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(adminchatID, "В этом чате пока нет пользователей.")
				b.bot.Send(msg)
			}
		} else {
			msg := tgbotapi.NewMessage(adminchatID, "У вас пока нет доступных чатов.")
			b.bot.Send(msg)
		}

	case strings.HasPrefix(query, "group "):
		groupName := strings.TrimPrefix(query, "group ")
		// Получаем все чаты пользователя
		chats := b.db.GetChatsForUser(userID)
		if len(chats) > 0 {
			// Используем первый чат из списка
			users := b.db.GetUsersForMention(chats[0].ChatID, groupName)
			if len(users) > 0 {
				msg := tgbotapi.NewMessage(adminchatID, strings.Join(users, " "))
				b.bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(adminchatID, "В этой группе пока нет пользователей.")
				b.bot.Send(msg)
			}
		} else {
			msg := tgbotapi.NewMessage(adminchatID, "У вас пока нет доступных чатов.")
			b.bot.Send(msg)
		}

	case strings.HasPrefix(query, "create_user"):
		msg := tgbotapi.NewMessage(adminchatID, "Введите данные пользователя в формате:\n/add_user <user_id> <username>")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Назад", "admin_users"),
			),
		)
		b.bot.Send(msg)

	case strings.HasPrefix(query, "create_chat"):
		msg := tgbotapi.NewMessage(adminchatID, "Введите данные чата в формате:\n/add_chat <chat_id> <title>")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Назад", "admin_chats"),
			),
		)
		b.bot.Send(msg)

	case strings.HasPrefix(query, "create_group"):
		msg := tgbotapi.NewMessage(adminchatID, "Введите название группы в формате:\n/add_group <name>")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Назад", "admin_groups"),
			),
		)
		b.bot.Send(msg)

	case strings.HasPrefix(query, "users_page_"):
		page, _ := strconv.Atoi(strings.TrimPrefix(query, "users_page_"))
		b.ShowUsersList(adminchatID, page, &update)

	case strings.HasPrefix(query, "chats_page_"):
		page, _ := strconv.Atoi(strings.TrimPrefix(query, "chats_page_"))
		b.ShowChatsList(adminchatID, page, &update)

	case strings.HasPrefix(query, "groups_page_"):
		page, _ := strconv.Atoi(strings.TrimPrefix(query, "groups_page_"))
		b.ShowGroupsList(adminchatID, page, &update)

	case strings.HasPrefix(query, "user_info_"):
		userID, _ := strconv.ParseInt(strings.TrimPrefix(query, "user_info_"), 10, 64)
		b.ShowUserInfo(adminchatID, userID)

	case strings.HasPrefix(query, "chat_info_"):
		OtherChatid, _ := strconv.ParseInt(strings.TrimPrefix(query, "chat_info_"), 10, 64)
		b.ShowChatInfo(adminchatID, OtherChatid)

	case strings.HasPrefix(query, "group_info_"):
		groupName := strings.TrimPrefix(query, "group_info_")
		b.ShowGroupInfo(adminchatID, groupName)

	case strings.HasPrefix(query, "edit_user_"):
		userID, _ := strconv.ParseInt(strings.TrimPrefix(query, "edit_user_"), 10, 64)
		// TODO: Реализовать редактирование пользователя
		msg := tgbotapi.NewMessage(adminchatID, fmt.Sprintf("Редактирование пользователя %d (в разработке)", userID))
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Назад", "admin_users"),
			),
		)
		b.bot.Send(msg)

	case strings.HasPrefix(query, "delete_user_"):
		userID, _ := strconv.ParseInt(strings.TrimPrefix(query, "delete_user_"), 10, 64)
		if err := b.db.DeleteUser(userID); err != nil {
			msg := tgbotapi.NewMessage(adminchatID, "Ошибка удаления пользователя")
			b.bot.Send(msg)
			return
		}
		msg := tgbotapi.NewMessage(adminchatID, "Пользователь успешно удален")
		b.bot.Send(msg)
		b.ShowAdminPanel(adminchatID)

	case strings.HasPrefix(query, "edit_chat_"):
		//chatID, _ := strconv.ParseInt(strings.TrimPrefix(query, "edit_chat_"), 10, 64)
		// TODO: Реализовать редактирование чата
		msg := tgbotapi.NewMessage(adminchatID, "Редактирование чата (в разработке)")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Назад", "admin_chats"),
			),
		)
		b.bot.Send(msg)

	case strings.HasPrefix(query, "delete_chat_"):
		chatID, _ := strconv.ParseInt(strings.TrimPrefix(query, "delete_chat_"), 10, 64)
		if err := b.db.DeleteChat(chatID); err != nil {
			msg := tgbotapi.NewMessage(adminchatID, "Ошибка удаления чата")
			b.bot.Send(msg)
			return
		}
		msg := tgbotapi.NewMessage(adminchatID, "Чат успешно удален")
		b.bot.Send(msg)
		b.ShowAdminPanel(adminchatID)

	case strings.HasPrefix(query, "edit_group_"):
		groupName := strings.TrimPrefix(query, "edit_group_")
		// TODO: Реализовать редактирование группы
		msg := tgbotapi.NewMessage(adminchatID, fmt.Sprintf("Редактирование группы %s (в разработке)", groupName))
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Назад", "admin_groups"),
			),
		)
		b.bot.Send(msg)

	case strings.HasPrefix(query, "delete_group_"):
		groupName := strings.TrimPrefix(query, "delete_group_")
		if err := b.db.DeleteGroup(groupName); err != nil {
			msg := tgbotapi.NewMessage(adminchatID, "Ошибка удаления группы")
			b.bot.Send(msg)
			return
		}
		msg := tgbotapi.NewMessage(adminchatID, "Группа успешно удалена")
		b.bot.Send(msg)
		b.ShowAdminPanel(adminchatID)

	case query == "admin_users":
		b.ShowUsersList(adminchatID, 1, &update)

	case query == "admin_chats":
		b.ShowChatsList(adminchatID, 1, &update)

	case query == "admin_groups":
		b.ShowGroupsList(adminchatID, 1, &update)

	case query == "admin_relations":
		b.ShowRelations(adminchatID)

	case query == "admin_view":
		b.ShowViewMenu(adminchatID)

	case query == "admin_back":
		b.ShowAdminPanel(adminchatID)

	case strings.HasPrefix(query, "add_users_to_chat_"):
		chatID, _ := strconv.ParseInt(strings.TrimPrefix(query, "add_users_to_chat_"), 10, 64)
		b.ShowUsersToAddToChat(adminchatID, chatID, 1, &update)

	case strings.HasPrefix(query, "add_users_to_chat_page_"):
		parts := strings.Split(strings.TrimPrefix(query, "add_users_to_chat_page_"), "_")
		chatID, _ := strconv.ParseInt(parts[0], 10, 64)
		page, _ := strconv.Atoi(parts[1])
		b.ShowUsersToAddToChat(adminchatID, chatID, page, &update)

	case strings.HasPrefix(query, "add_user_to_chat_"):
		parts := strings.Split(strings.TrimPrefix(query, "add_user_to_chat_"), "_")
		chatID, _ := strconv.ParseInt(parts[0], 10, 64)
		userID, _ := strconv.ParseInt(parts[1], 10, 64)

		if err := b.db.AddUserToChat(userID, chatID); err != nil {
			msg := tgbotapi.NewMessage(adminchatID, fmt.Sprintf("Ошибка добавления пользователя в чат: %v", err))
			b.bot.Send(msg)
			return
		}

		// Получаем информацию о пользователе для сообщения
		user, err := b.db.GetUser(userID)
		if err != nil {
			msg := tgbotapi.NewMessage(adminchatID, "Пользователь успешно добавлен в чат")
			b.bot.Send(msg)
			return
		}

		msg := tgbotapi.NewMessage(adminchatID, fmt.Sprintf("Пользователь @%s успешно добавлен в чат", user.Username))
		b.bot.Send(msg)
		b.ShowAdminPanel(adminchatID)

	case strings.HasPrefix(query, "add_users_to_group_"):
		groupName := strings.TrimPrefix(query, "add_users_to_group_")
		b.ShowUsersToAddToGroup(adminchatID, groupName, 1, &update)

	case strings.HasPrefix(query, "add_users_to_group_page_"):
		parts := strings.Split(strings.TrimPrefix(query, "add_users_to_group_page_"), "_")
		groupName := parts[0]
		page, _ := strconv.Atoi(parts[1])
		b.ShowUsersToAddToGroup(adminchatID, groupName, page, &update)

	case strings.HasPrefix(query, "add_user_to_group_"):
		parts := strings.Split(strings.TrimPrefix(query, "add_user_to_group_"), "_")
		groupName := parts[0]
		userID, _ := strconv.ParseInt(parts[1], 10, 64)

		if err := b.db.AddUsersToGroup([]int64{userID}, groupName); err != nil {
			msg := tgbotapi.NewMessage(adminchatID, fmt.Sprintf("Ошибка добавления пользователя в группу: %v", err))
			b.bot.Send(msg)
			return
		}

		// Получаем информацию о пользователе для сообщения
		user, err := b.db.GetUser(userID)
		if err != nil {
			msg := tgbotapi.NewMessage(adminchatID, "Пользователь успешно добавлен в группу")
			b.bot.Send(msg)
			return
		}

		msg := tgbotapi.NewMessage(adminchatID, fmt.Sprintf("Пользователь @%s успешно добавлен в группу", user.Username))
		b.bot.Send(msg)
		b.ShowAdminPanel(adminchatID)

	default:
		msg := tgbotapi.NewMessage(adminchatID, "Неизвестное действие.")
		b.bot.Send(msg)
	}
}

func (b *TelegramBot) HandleInlineQuery(update tgbotapi.Update) {
	query := update.InlineQuery
	if query == nil {
		return
	}

	// Показываем список групп для пользователя
	var results []interface{}
	
	// Получаем группы для пользователя
	userGroups := b.db.GetGroupsForUser(query.From.ID)
	
	// Получаем текущий чат пользователя
	chats := b.db.GetChatsForUser(query.From.ID)
	var currentChatID int64
	if len(chats) > 0 {
		currentChatID = chats[0].ChatID
	}
	
	if len(userGroups) > 0 {
		for _, group := range userGroups {
			groupButton := tgbotapi.NewInlineQueryResultArticle(
				query.ID+"_"+group.Name,
				fmt.Sprintf("Группа: %s", group.Name),
				fmt.Sprintf("Упомянуть пользователей группы %s", group.Name),
			)
			
			// Получаем список пользователей группы
			var groupMentionText string
			if currentChatID != 0 {
				groupUsers := b.db.GetUsersForMention(currentChatID, group.Name)
				if len(groupUsers) > 0 {
					groupMentionText = strings.Join(groupUsers, " ")
				} else {
					groupMentionText = fmt.Sprintf("В группе %s нет пользователей в текущем чате.", group.Name)
				}
			} else {
				groupMentionText = fmt.Sprintf("Не удалось определить текущий чат. Используйте команду /group %s в нужном чате.", group.Name)
			}
			
			groupButton.Description = fmt.Sprintf("Отметить участников группы %s в чате", group.Name)
			groupButton.InputMessageContent = tgbotapi.InputTextMessageContent{
				Text: groupMentionText,
			}
			results = append(results, groupButton)
		}
	} else {
		// Если у пользователя нет групп
		noGroups := tgbotapi.NewInlineQueryResultArticle(
			query.ID+"_no_groups",
			"У вас нет доступных групп",
			"Нет доступных групп",
		)
		noGroups.Description = "Обратитесь к администратору для добавления в группы"
		noGroups.InputMessageContent = tgbotapi.InputTextMessageContent{
			Text: "У вас нет доступных групп. Обратитесь к администратору для добавления в группы.",
		}
		results = append(results, noGroups)
	}
	
	// Добавляем информационную кнопку про упоминание всех пользователей
	infoButton := tgbotapi.NewInlineQueryResultArticle(
		query.ID+"_info",
		"Для упоминания всех используйте /all",
		"Информация",
	)
	infoButton.Description = "Чтобы упомянуть всех участников чата, используйте команду /all или /everyone в чате"
	infoButton.InputMessageContent = tgbotapi.InputTextMessageContent{
		Text: "Чтобы упомянуть всех участников чата, используйте команду /all или /everyone непосредственно в чате.",
	}
	infoButton.ThumbURL = "https://img.icons8.com/color/48/000000/info.png"
	results = append(results, infoButton)
	
	// Отправляем результаты
	inlineConfig := tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		Results:       results,
		CacheTime:     0,
	}
	b.bot.Send(inlineConfig)
}
