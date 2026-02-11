package bot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"era_sporta_bot_ruletka/internal/domain"
	"era_sporta_bot_ruletka/internal/service"
	"era_sporta_bot_ruletka/internal/telegram"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
)

const (
	msgSubscribe     = "Привет! 👋 Добро пожаловать в Колесо Фортуны от фитнес-клуба «Эра Спорта».\n\nЧтобы крутить рулетку 🎯, нужно выполнить два простых действия.\n\nШаг 1 — подписаться на наш официальный Telegram-канал 🔔\nТам мы публикуем новости клуба, предложения и полезную информацию 💪\n\nКак будете готовы, нажмите кнопку «Я подписался» 👇"
	msgShareOfficial = "Шаг 2 — номер телефона 📱\n\nНомер нужен, чтобы наш менеджер мог\nсвязаться с вами и подтвердить результат.\n\nМы используем только официальный способ Telegram\nи не передаём номер третьим лицам 🤝\n\nНажмите «Поделиться номером» ниже 👇"
	msgPhoneSaved    = "✅ Отлично! Номер сохранён. Нажмите кнопку ниже, чтобы открыть приложение и крутить рулетку."
	msgWelcomeBack   = "👋 С возвращением! Нажмите кнопку ниже, чтобы открыть приложение."
)

type Handler struct {
	bot        *tgbotapi.BotAPI
	userSvc    *service.UserService
	notifier   *Notifier
	webAppURL  string
	channelID  int64
	channelURL string
}

func NewHandler(bot *tgbotapi.BotAPI, userSvc *service.UserService, notifier *Notifier, webAppURL string, channelID int64, channelURL string) *Handler {
	return &Handler{
		bot:        bot,
		userSvc:    userSvc,
		notifier:   notifier,
		webAppURL:  webAppURL,
		channelID:  channelID,
		channelURL: channelURL,
	}
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
			if h.channelURL == "" {
				h.send(chatID, "Канал для подписки не настроен. Напишите администратору.")
				return
			}
			msg := tgbotapi.NewMessage(chatID, msgSubscribe)
			msg.ReplyMarkup = SubscribeInlineMarkup(h.channelURL)
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
		// Already has phone — показываем кнопку открытия приложения
		h.sendAppCard(chatID)
		return
	}

	// Need phone — приветствие с inline-кнопкой
	if h.channelURL == "" {
		h.send(chatID, "Канал для подписки не настроен. Напишите администратору.")
		return
	}
	msg := tgbotapi.NewMessage(chatID, msgSubscribe)
	msg.ReplyMarkup = SubscribeInlineMarkup(h.channelURL)
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("[bot] Send error: %v", err)
	}
}

func (h *Handler) handleCallback(ctx context.Context, q *tgbotapi.CallbackQuery) {
	switch q.Data {
	case "check_subscribe":
		if h.channelID == 0 || h.channelURL == "" {
			h.send(q.Message.Chat.ID, "Канал для подписки не настроен. Напишите администратору.")
			break
		}
		member, err := telegram.IsUserMember(ctx, h.bot.Token, h.channelID, q.From.ID)
		if err != nil || !member {
			msg := tgbotapi.NewMessage(q.Message.Chat.ID, "Похоже, вы ещё не подписались. Подпишитесь и нажмите «Я подписался» ещё раз.")
			msg.ReplyMarkup = SubscribeInlineMarkup(h.channelURL)
			if _, sendErr := h.bot.Send(msg); sendErr != nil {
				log.Printf("[bot] Send error: %v", sendErr)
			}
			break
		}
		msg := tgbotapi.NewMessage(q.Message.Chat.ID, msgShareOfficial)
		msg.ReplyMarkup = SharePhoneKeyboard()
		if _, err := h.bot.Send(msg); err != nil {
			log.Printf("[bot] Send error: %v", err)
		}
	case "share_phone":
		if h.channelID != 0 {
			member, err := telegram.IsUserMember(ctx, h.bot.Token, h.channelID, q.From.ID)
			if err != nil || !member {
				if h.channelURL != "" {
					msg := tgbotapi.NewMessage(q.Message.Chat.ID, msgSubscribe)
					msg.ReplyMarkup = SubscribeInlineMarkup(h.channelURL)
					if _, sendErr := h.bot.Send(msg); sendErr != nil {
						log.Printf("[bot] Send error: %v", sendErr)
					}
				} else {
					h.send(q.Message.Chat.ID, "Канал для подписки не настроен. Напишите администратору.")
				}
				break
			}
		}
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
	// Then show Open App button
	h.sendAppCard(chatID)
}

func (h *Handler) send(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("[bot] Send error: %v", err)
	}
}

func getPromoImagePath() string {
	name := "wheel_promo.png"
	// 1) Переменная окружения
	if p := os.Getenv("PROMO_IMAGE_PATH"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	// 2) Относительно текущей директории (запуск из корня проекта)
	cwd, _ := os.Getwd()
	for _, dir := range []string{cwd, "/root/era_sporta_bot_ruletka"} {
		p := filepath.Join(dir, "assets", name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func (h *Handler) sendAppCard(chatID int64) {
	imgPath := getPromoImagePath()
	if imgPath != "" {
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(imgPath))
		photo.Caption = msgWelcomeBack
		photo.ReplyMarkup = OpenAppKeyboard(h.webAppURL)
		if _, err := h.bot.Send(photo); err != nil {
			log.Printf("[bot] Send photo error: %v", err)
		} else {
			return
		}
	}
	// Только текст и кнопка (если фото нет или отправка не удалась)
	msg := tgbotapi.NewMessage(chatID, msgWelcomeBack)
	msg.ReplyMarkup = OpenAppKeyboard(h.webAppURL)
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
