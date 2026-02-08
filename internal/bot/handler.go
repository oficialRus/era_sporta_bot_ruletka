package bot

import (
	"context"
	"errors"
	"fmt"
	"log"

	"era_sporta_bot_ruletka/internal/domain"
	"era_sporta_bot_ruletka/internal/service"

	"github.com/jackc/pgx/v5"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	msgNeedPhone     = "👋 Добро пожаловать! Чтобы получить доступ к рулетке с бонусами, нажмите кнопку ниже."
	msgShareOfficial = "Нажмите кнопку ниже, чтобы поделиться номером из вашего аккаунта Telegram. Принимается только официальный контакт."
	msgPhoneSaved    = "✅ Отлично! Номер сохранён. Нажмите кнопку ниже, чтобы открыть приложение и крутить рулетку."
	msgWelcomeBack   = "👋 С возвращением! Нажмите кнопку ниже, чтобы открыть приложение."
	msgOpenLocalLink = "Откройте приложение по ссылке (локальная разработка):"
)

type Handler struct {
	bot       *tgbotapi.BotAPI
	userSvc   *service.UserService
	notifier  *Notifier
	webAppURL string
}

func NewHandler(bot *tgbotapi.BotAPI, userSvc *service.UserService, notifier *Notifier, webAppURL string) *Handler {
	return &Handler{bot: bot, userSvc: userSvc, notifier: notifier, webAppURL: webAppURL}
}

func (h *Handler) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
	// Inline-кнопка «Поделиться номером»
	if update.CallbackQuery != nil {
		h.handleCallback(ctx, update.CallbackQuery)
		return
	}
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

	// Контакт из Telegram (только так принимаем номер — подделать нельзя)
	if msg.Contact != nil {
		h.handleContact(ctx, chatID, msg.From, msg.Contact)
		return
	}
}

func (h *Handler) handleStart(ctx context.Context, chatID int64, from *tgbotapi.User) {
	user, err := h.userSvc.GetByTelegramID(ctx, from.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Новый пользователь — приветствие с inline-кнопкой
			msg := tgbotapi.NewMessage(chatID, msgNeedPhone)
			msg.ReplyMarkup = SharePhoneInlineMarkup()
			if _, sendErr := h.bot.Send(msg); sendErr != nil {
				log.Printf("[bot] Send error: %v", sendErr)
			}
			return
		}
		log.Printf("[bot] GetByTelegramID error: %v", err)
		h.send(chatID, "Произошла ошибка. Попробуйте позже.")
		return
	}

	if user != nil && user.Phone != "" {
		// Already has phone — кнопка или ссылка (localhost в кнопке Telegram не принимает)
		msg := tgbotapi.NewMessage(chatID, msgWelcomeBack)
		if !IsLocalhostURL(h.webAppURL) {
			msg.ReplyMarkup = OpenAppKeyboard(h.webAppURL)
		} else {
			// Ссылка в тексте кликабельна в Telegram Desktop — откроется локально
			msg.Text = msgOpenLocalLink + "\n" + h.webAppURL
		}
		if _, err := h.bot.Send(msg); err != nil {
			log.Printf("[bot] Send error: %v", err)
		}
		return
	}

	// Need phone — приветствие с inline-кнопкой
	msg := tgbotapi.NewMessage(chatID, msgNeedPhone)
	msg.ReplyMarkup = SharePhoneInlineMarkup()
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("[bot] Send error: %v", err)
	}
}

func (h *Handler) handleCallback(_ context.Context, q *tgbotapi.CallbackQuery) {
	if q.Data == "share_phone" {
		// Показываем только официальную кнопку Telegram «Поделиться контактом» — номер подделать нельзя
		msg := tgbotapi.NewMessage(q.Message.Chat.ID, msgShareOfficial)
		msg.ReplyMarkup = SharePhoneKeyboard()
		if _, err := h.bot.Send(msg); err != nil {
			log.Printf("[bot] Send error: %v", err)
		}
	}
	if _, err := h.bot.Request(tgbotapi.NewCallback(q.ID, "")); err != nil {
		log.Printf("[bot] Answer callback error: %v", err)
	}
}

func (h *Handler) handleContact(ctx context.Context, chatID int64, from *tgbotapi.User, contact *tgbotapi.Contact) {
	// Принимаем только контакт от самого пользователя (номер из аккаунта Telegram, подделать нельзя)
	if contact.UserID != 0 && contact.UserID != from.ID {
		h.send(chatID, "Пожалуйста, нажмите «Поделиться контактом» и отправьте именно свой номер из Telegram.")
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

	// Уведомление в админский чат о новом пользователе
	h.notifyNewUser(ctx, phone, from)

	// Remove reply keyboard first
	rmMsg := tgbotapi.NewMessage(chatID, msgPhoneSaved)
	rmMsg.ReplyMarkup = RemoveKeyboard()
	if _, err := h.bot.Send(rmMsg); err != nil {
		log.Printf("[bot] Send error: %v", err)
		return
	}
	// Then show Open App button or clickable localhost link
	appMsg := tgbotapi.NewMessage(chatID, "Нажмите кнопку ниже:")
	if !IsLocalhostURL(h.webAppURL) {
		appMsg.ReplyMarkup = OpenAppKeyboard(h.webAppURL)
	} else {
		appMsg.Text = msgOpenLocalLink + "\n" + h.webAppURL
	}
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

// notifyNewUser отправляет в админский чат сообщение о пользователе, который поделился номером.
func (h *Handler) notifyNewUser(ctx context.Context, phone string, from *tgbotapi.User) {
	if h.notifier == nil {
		return
	}
	name := from.FirstName
	if from.LastName != "" {
		name += " " + from.LastName
	}
	if name == "" && from.UserName != "" {
		name = "@" + from.UserName
	}
	if name == "" {
		name = "—"
	}
	text := fmt.Sprintf("Новый пользователь:\nНомер - %s\nИмя - %s\nId - %d", phone, name, from.ID)
	h.notifier.Notify(ctx, text)
}

// NotifyAdmin sends spin notification to admin
func (h *Handler) NotifyAdmin(ctx context.Context, user *domain.User, prizeName string) {
	if h.notifier != nil {
		text := fmt.Sprintf("🎰 Новый спин!\nПользователь: %s %s (@%s)\nТелефон: %s\nПриз: %s",
			user.FirstName, user.LastName, user.Username, user.Phone, prizeName)
		h.notifier.NotifyWithTime(ctx, text)
	}
}
