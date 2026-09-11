package handlers

import (
	"bankTg/intern/domain"
	"bankTg/intern/service"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TgBotHandler struct {
	service    *service.Service
	bot        *tgbotapi.BotAPI
	userStates map[int64]string
	rentData   map[int64]struct {
		Service string
		Amount  int
		Hours   int
	}
}

func CreateTgBotHandler(service *service.Service,
	bot *tgbotapi.BotAPI) *TgBotHandler {
	return &TgBotHandler{
		service: service,
		bot:     bot,
	}
}
func (t *TgBotHandler) HandleUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		t.handleMessage(update.Message)
		return
	}
	if update.CallbackQuery != nil {
		t.handleCallBack(*update.CallbackQuery)
		return
	}

	slog.Warn("Неизвестный тип апдейта", "update", update)
}
func (t *TgBotHandler) handleCallBack(query tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID
	data := query.Data
	t.bot.Send(tgbotapi.NewCallback(query.ID, ""))
	if strings.HasPrefix(data, "service_") {
		serviceMap := map[string]string{
			"service_telegram":  "Telegram",
			"service_yandex":    "Yandex",
			"service_mailru":    "Mail.ru",
			"service_vk":        "VK",
			"service_avito":     "Avito",
			"service_hhru":      "hh.ru",
			"service_ozon":      "ozon",
			"service_wb":        "wildberries",
			"service_instagram": "Instagram",
			"service_whatsapp":  "WhatsApp",
			"service_google":    "Google",
			"service_facebook":  "FaceBook",
			"service_other":     "Другой",
		}

		service, ok := serviceMap[data]
		if !ok {
			t.sendMessage(chatID, "❌ Неизвестный сервис")
			return
		}
		rentData := t.rentData[chatID]
		rentData.Service = service
		t.rentData[chatID] = rentData
		t.userStates[chatID] = "awaiting_amount"
		t.sendMessage(chatID, fmt.Sprintf("✅ Сервис: %s\n\n📞 Введите количество SIM-карт:", service))
		return
	}
	switch data {
	case "action_cards":
		username := query.From.UserName
		if username == "" {
			username = query.From.FirstName
		}
		t.sendMessage(chatID, fmt.Sprintf("📋 *%s*, вот твои карты: ...", username)) // ЗАГЛУШКА
	case "action_balance":
		username := query.From.UserName
		if username == "" {
			username = query.From.FirstName
		}
		t.sendMessage(chatID, fmt.Sprintf("💰 *%s*, твой баланс: 1000 руб.", username)) // ЗАГЛУШКА
	case "lang_ru":
		t.sendMessage(chatID, "🇷🇺 *Язык установлен: Русский*")
		t.sendMainKeyboard(chatID) // ← Главное меню

	case "lang_en":
		t.sendMessage(chatID, "🇬🇧 *Language set: English*")
		t.sendMainKeyboard(chatID)

	case "lang_ua":
		t.sendMessage(chatID, "🇺🇦 *Мову встановлено: Українська*")
		t.sendMainKeyboard(chatID)
	default:
		t.sendMessage(chatID, "Неизвестное действие")
	}
}
func (t *TgBotHandler) sendMessage(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	if _, err := t.bot.Send(msg); err != nil {
		slog.Error("Ошибка отправки", "error", err)
	}
}
func (t *TgBotHandler) handleMessage(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	text := message.Text
	if state, ok := t.userStates[chatID]; ok {
		t.handleRentFlow(chatID, message.From, state, text)
		return
	}
	if message.IsCommand() {
		switch message.Command() {
		case "start":
			t.handleStart(chatID, message.From)
		case "cards":
			t.handleCards(chatID, message.From)
		case "balance":
			t.handleBalance(chatID, message.From)
		case "help":
			t.handleHelp(chatID)
		case "rent":
			t.handleRent(chatID, message.From, message.Text)
		default:
			t.sendMessage(chatID, "❌ Неизвестная команда. Напиши /help")
		}
		return
	}

	// Обработка нажатий на ReplyKeyboard (кнопки меню)
	switch message.Text {
	case "📋 Мои карты":
		t.handleCards(chatID, message.From)
	case "💰 Баланс":
		t.handleBalance(chatID, message.From)
	case "🆘 Помощь":
		t.handleHelp(chatID)
	case "🔗Аренда сим-карты":
		t.startRentFlow(chatID)
	default:
		t.sendMessage(chatID, "Я понимаю только команды и кнопки. Напиши /help")
	}
}
func (t *TgBotHandler) handleStart(chatID int64, user *tgbotapi.User) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	username := getUserName(user)
	telegramID := user.ID

	// 1. Пытаемся зарегистрировать пользователя
	err := t.service.CreateUser(ctx, telegramID, user.UserName)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			// ✅ Если пользователь уже есть
			t.sendMessage(chatID, fmt.Sprintf("👋 *И снова здравствуй, %s!*", username))
			t.sendLanguageKeyboard(chatID) // ← Отправляем выбор языка
			return
		}
		slog.Error("Failed to register user", "telegram_id", telegramID, "error", err)
		t.sendMessage(chatID, "❌ Ошибка регистрации. Попробуйте позже.")
		return
	}

	// ✅ Если пользователь новый
	t.sendMessage(chatID, fmt.Sprintf("🎉 *Добро пожаловать, %s!*", username))
	t.sendLanguageKeyboard(chatID)
}
func (t *TgBotHandler) handleBalance(chatID int64, user *tgbotapi.User) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	balance, err := t.service.GetBalance(ctx, user.ID)
	if err != nil {
		t.sendMessage(chatID, "❌ Ошибка получения баланса")
		return
	}

	username := getUserName(user)
	text := fmt.Sprintf("💰 *%s*, ваш баланс: *%.2f USD*", username, balance)
	t.sendMessage(chatID, text)
}
func (t *TgBotHandler) handleCards(chatID int64, user *tgbotapi.User) {
	// 1. Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 2. Получаем список активных карт пользователя
	cards, err := t.service.GetActiveCards(ctx, user.ID)
	if err != nil {
		// Обрабатываем ошибки
		if errors.Is(err, domain.ErrUserNotFound) {
			t.sendMessage(chatID, "❌ Вы не зарегистрированы. Напишите /start")
			return
		}
		if errors.Is(err, domain.ErrNoActiveCards) {
			t.sendMessage(chatID, "📭 У вас пока нет активных карт.")
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			t.sendMessage(chatID, "⏳ Сервер не отвечает. Попробуйте позже.")
			return
		}
		// Другие ошибки
		slog.Error("GetActiveCards error", "telegram_id", user.ID, "error", err)
		t.sendMessage(chatID, "❌ Ошибка получения карт. Попробуйте позже.")
		return
	}

	// 3. Если карт нет — сообщаем об этом
	if len(cards) == 0 {
		t.sendMessage(chatID, "📭 У вас пока нет активных карт.")
		return
	}

	// 4. Форматируем список карт
	username := getUserName(user)
	var result strings.Builder
	result.WriteString(fmt.Sprintf("📋 *%s*, ваши активные карты:*\n\n", username))

	for i, card := range cards {
		// Форматируем дату окончания аренды (если есть)
		expiresAt := "без ограничений"
		if !card.ExpiresAt.IsZero() {
			expiresAt = card.ExpiresAt.Format("02.01.2006 15:04")
		}

		// Номер телефона маскируем (показываем только последние 4 цифры)
		maskedPhone := maskPhone(card.Phone)

		result.WriteString(fmt.Sprintf(
			"%d️⃣ 🆔 *%d* | 📱 `%s`\n",
			i+1,
			card.Id,
			maskedPhone,
		))

		if expiresAt != "без ограничений" {
			result.WriteString(fmt.Sprintf("   ⏳ До: *%s*\n", expiresAt))
		}

		if card.Status == "busy" {
			result.WriteString("   🔴 *Занята*\n")
		} else {
			result.WriteString("   🟢 *Свободна*\n")
		}
		result.WriteString("\n")
	}

	// 5. Отправляем ответ
	t.sendMessage(chatID, result.String())
}

// getUserName возвращает красивое имя пользователя для отображения
// Приоритет: @username → FirstName → "Пользователь"
func getUserName(user *tgbotapi.User) string {
	// 1. Если есть @username — используем его
	if user.UserName != "" {
		return "@" + user.UserName
	}

	// 2. Если есть FirstName — используем его
	if user.FirstName != "" {
		return user.FirstName
	}

	// 3. Если вообще ничего нет — возвращаем "Пользователь"
	return "Пользователь"
}
func (t *TgBotHandler) handleHelp(chatID int64) {
	text := `
📋 Команды:
/start — начать работу и зарегистрироваться
/cards — показать список ваших карт
/balance — проверить баланс
/help — показать эту справку

📌 Кнопки:
Внизу экрана есть меню с кнопками для быстрого доступа.

💡 Пример использования:
1. Напишите /start для регистрации
2. Используйте /cards или кнопку "Мои карты", чтобы посмотреть доступные SIM-карты
3. Проверьте баланс через /balance или кнопку "Баланс"

🆘 *Если что-то пошло не так:* напиши support'у - ссылка на него в шапке бота!

🚀 Приятного использования!
`
	t.sendMessage(chatID, text)
}

// sendMainKeyboard отправляет основное меню с кнопками (ReplyKeyboard)
func (t *TgBotHandler) sendMainKeyboard(chatID int64) {
	// 1. Создаем клавиатуру с кнопками
	keyboard := tgbotapi.NewReplyKeyboard(
		// Первый ряд (2 кнопки)
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📋 Мои карты"),
			tgbotapi.NewKeyboardButton("💰 Баланс"),
		),
		// Второй ряд (1 кнопка)
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🆘 Помощь"),
			tgbotapi.NewKeyboardButton("🔗Аренда сим-карты"),
		),
	)

	// 2. Создаем сообщение с клавиатурой
	msg := tgbotapi.NewMessage(chatID, "👇 Выберите действие:")
	msg.ReplyMarkup = keyboard

	// 3. Отправляем
	if _, err := t.bot.Send(msg); err != nil {
		slog.Error("Ошибка отправки клавиатуры", "chat_id", chatID, "error", err)
	}
}

// maskPhone маскирует номер телефона, показывая только последние 4 цифры
// Пример: +79001234567 → ****4567
// Пример: 89001234567 → ****4567
// Пример: 1234 → 1234 (если номер короткий, не маскируем)
func maskPhone(phone string) string {
	// 1. Если номер пустой — возвращаем как есть
	if phone == "" {
		return phone
	}

	// 2. Если номер слишком короткий (< 4 цифр) — не маскируем
	if len(phone) < 4 {
		return phone
	}

	// 3. Показываем последние 4 цифры
	return "****" + phone[len(phone)-4:]
}

// sendLanguageKeyboard отправляет клавиатуру для выбора языка
func (t *TgBotHandler) sendLanguageKeyboard(chatID int64) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🇷🇺 Русский", "lang_ru"),
			tgbotapi.NewInlineKeyboardButtonData("🇬🇧 English", "lang_en"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🇺🇦 Українська", "lang_ua"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "👇 Выберите язык / Choose language:")
	msg.ReplyMarkup = keyboard
	t.bot.Send(msg)
}

// handlers/telegram.go

func (t *TgBotHandler) handleRent(chatID int64, user *tgbotapi.User, text string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	data, ok := t.rentData[chatID]
	if !ok {
		t.sendMessage(chatID, "❌ Сначала выберите сервис через меню аренды.")
		return
	}

	service := data.Service
	if service == "" {
		t.sendMessage(chatID, "❌ Сервис не выбран. Начните аренду заново.")
		return
	}
	_, err := t.service.RentCard(ctx, user.ID, service)
	if err != nil {
		// Обрабатываем ошибки
		if strings.Contains(err.Error(), "no free cards") {
			t.sendMessage(chatID, "❌ Нет свободных SIM-карт.")
			return
		}
		if strings.Contains(err.Error(), "insufficient balance") {
			t.sendMessage(chatID, "❌ Недостаточно средств.")
			return
		}
		slog.Error("RentCard error", "telegram_id", user.ID, "error", err)
		t.sendMessage(chatID, "❌ Ошибка аренды. Попробуйте позже.")
		return
	}

	username := getUserName(user)
	text = fmt.Sprintf(
		"✅ *%s*, вы арендовали сим-карту на 20 минут!",
		username,
	)
	t.sendMessage(chatID, text)
}

// sendServiceKeyboard — отправляет выбор сервиса
func (t *TgBotHandler) sendServiceKeyboard(chatID int64) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Telegram", "service_telegram"),
			tgbotapi.NewInlineKeyboardButtonData("Mail.ru", "service_mailru"),
			tgbotapi.NewInlineKeyboardButtonData("FaceBook", "service_facebook"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Instagram", "service_instagram"),
			tgbotapi.NewInlineKeyboardButtonData("WhatsApp", "service_whatsapp"),
			tgbotapi.NewInlineKeyboardButtonData("Yandex", "service_yandex"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Google", "service_google"),
			tgbotapi.NewInlineKeyboardButtonData("VK", "service_vk"),
			tgbotapi.NewInlineKeyboardButtonData("Avito", "service_avito"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("hh.ru", "service_hhru"),
			tgbotapi.NewInlineKeyboardButtonData("ozon", "service_ozon"),
			tgbotapi.NewInlineKeyboardButtonData("wildberries", "service_wb"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Другое", "service_other"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "📝 Выберите сервис, для которого нужны SIM-карты:")
	msg.ReplyMarkup = keyboard
	t.bot.Send(msg)
}

// startRentFlow — начало процесса аренды (показываем сервисы)
func (t *TgBotHandler) startRentFlow(chatID int64) {
	t.sendServiceKeyboard(chatID)
}
func (t *TgBotHandler) handleRentFlow(chatID int64, user *tgbotapi.User, state, text string) {
	switch state {
	case "awaiting_amount":
		amount, err := strconv.Atoi(text)
		if err != nil || amount <= 0 {
			t.sendMessage(chatID, "❌ Введите положительное число. Например: 2")
			return
		}

		data := t.rentData[chatID]
		data.Amount = amount
		t.rentData[chatID] = data

		t.userStates[chatID] = "awaiting_hours"
		t.sendMessage(chatID, "⏳ Введите количество часов (максимум 72):")

	case "awaiting_hours":
		hours, err := strconv.Atoi(text)
		if err != nil || hours <= 0 {
			t.sendMessage(chatID, "❌ Введите положительное число. Например: 24")
			return
		}
		if hours > 72 {
			t.sendMessage(chatID, "❌ Максимум 72 часа.")
			return
		}

		data := t.rentData[chatID]
		data.Hours = hours
		t.rentData[chatID] = data

		// 🚀 Аренда
		t.executeRent(chatID, user, data.Service, data.Amount, data.Hours)

		// Очищаем состояние
		delete(t.userStates, chatID)
		delete(t.rentData, chatID)
	}
}
func (t *TgBotHandler) executeRent(chatID int64, user *tgbotapi.User, service string, amount, hours int) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rentedCount, err := t.service.RentCard(ctx, user.ID, service)
	if err != nil {
		if strings.Contains(err.Error(), "no free cards") {
			t.sendMessage(chatID, "❌ Нет свободных SIM-карт.")
			return
		}
		if strings.Contains(err.Error(), "insufficient balance") {
			t.sendMessage(chatID, "❌ Недостаточно средств.")
			return
		}
		slog.Error("RentCard error", "telegram_id", user.ID, "error", err)
		t.sendMessage(chatID, "❌ Ошибка аренды. Попробуйте позже.")
		return
	}

	username := getUserName(user)
	text := fmt.Sprintf(
		"✅ *%s*, вы арендовали *%d* карт на *%d* часов!\n📝 Сервис: %s",
		username,
		rentedCount,
		hours,
		service,
	)
	t.sendMessage(chatID, text)
}
