package bot

import (
	"context"
	"fmt"
	"log"

	"era_sporta_bot_ruletka/internal/domain"
	"era_sporta_bot_ruletka/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	msgNeedPhone   = "👋 Добро пожаловать! Чтобы получить доступ к рулетке с бонусами, поделитесь номером телефона."
	msgPhoneSaved  = "✅ Отлично! Номер сохранён. Нажмите кнопку ниже, чтобы открыть приложение и крутить рулетку."
	msgWelcomeBack = "👋 С возвращением! Нажмите кнопку ниже, чтобы открыть приложение."
)

type Handler struct {
	bot        *tgbotapi.BotAPI
	userSvc    *service.UserService
	notifier   *Notifier
	webAppURL  string
}

func NewHandler(bot *tgbotapi.BotAPI, userSvc *service.UserService, notifier *Notifier, webAppURL string) *Handler {
	return &Handler{bot: bot, userSvc: userSvc, notifier: notifier, webAppURL: webAppURL}
}

func (h *Handler) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	msg := update.Message
	chatID := msg.Chat.ID

	// /start
	if msg.IsCommand() && msg.Command() == "start" {
		h.handleStart(ctx, chatID, msg.From)
		return
	}

	// Contact shared
	if msg.Contact != nil {
		h.handleContact(ctx, chatID, msg.From, msg.Contact)
		return
	}
}

func (h *Handler) handleStart(ctx context.Context, chatID int64, from *tgbotapi.User) {
	user, err := h.userSvc.GetByTelegramID(ctx, from.ID)
	if err != nil {
		log.Printf("[bot] GetByTelegramID error: %v", err)
		h.send(chatID, "Произошла ошибка. Попробуйте позже.")
		return
	}

	if user != nil && user.Phone != "" {
		// Already has phone - show Open App button
		msg := tgbotapi.NewMessage(chatID, msgWelcomeBack)
		msg.ReplyMarkup = OpenAppKeyboard(h.webAppURL)
		if _, err := h.bot.Send(msg); err != nil {
			log.Printf("[bot] Send error: %v", err)
		}
		return
	}

	// Need phone
	msg := tgbotapi.NewMessage(chatID, msgNeedPhone)
	msg.ReplyMarkup = SharePhoneKeyboard()
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("[bot] Send error: %v", err)
	}
}

func (h *Handler) handleContact(ctx context.Context, chatID int64, from *tgbotapi.User, contact *tgbotapi.Contact) {
	// Only accept contact from the same user
	if contact.UserID != from.ID {
		h.send(chatID, "Пожалуйста, поделитесь своим номером телефона.")
		return
	}

	phone := normalizePhone(contact.PhoneNumber)
	if phone == "" {
		h.send(chatID, "Не удалось распознать номер. Попробуйте ещё раз.")
		return
	}

	user := &domain.User{
		TelegramUserID: int64(from.ID),
		Phone:          phone,
		FirstName:      from.FirstName,
		LastName:       from.LastName,
		Username:       from.UserName,
	}

	if err := h.userSvc.Upsert(ctx, user); err != nil {
		log.Printf("[bot] Upsert user error: %v", err)
		h.send(chatID, "Не удалось сохранить номер. Попробуйте позже.")
		return
	}

	// Remove reply keyboard first
	rmMsg := tgbotapi.NewMessage(chatID, msgPhoneSaved)
	rmMsg.ReplyMarkup = RemoveKeyboard()
	if _, err := h.bot.Send(rmMsg); err != nil {
		log.Printf("[bot] Send error: %v", err)
		return
	}
	// Then show Open App button
	appMsg := tgbotapi.NewMessage(chatID, "Нажмите кнопку ниже:")
	appMsg.ReplyMarkup = OpenAppKeyboard(h.webAppURL)
	if _, err := h.bot.Send(appMsg); err != nil {
		log.Printf("[bot] Send error: %v", err)
	}
}

func (h *Handler) send(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("[bot] Send error: %v", err)
	}
}

func normalizePhone(phone string) string {
	// Remove spaces, dashes, keep + and digits
	var b []byte
	for _, r := range phone {
		if r >= '0' && r <= '9' || r == '+' {
			b = append(b, byte(r))
		}
	}
	return string(b)
}

// NotifyAdmin sends spin notification to admin
func (h *Handler) NotifyAdmin(ctx context.Context, user *domain.User, prizeName string) {
	if h.notifier != nil {
		text := fmt.Sprintf("🎰 Новый спин!\nПользователь: %s %s (@%s)\nТелефон: %s\nПриз: %s",
			user.FirstName, user.LastName, user.Username, user.Phone, prizeName)
		h.notifier.NotifyWithTime(ctx, text)
	}
}
