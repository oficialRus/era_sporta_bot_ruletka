package bot

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// IsLocalhostURL возвращает true, если URL — localhost/127.0.0.1. Telegram не принимает такие URL в inline-кнопках.
func IsLocalhostURL(url string) bool {
	u := strings.ToLower(strings.TrimSpace(url))
	return strings.HasPrefix(u, "http://localhost") || strings.HasPrefix(u, "https://localhost") ||
		strings.Contains(u, "127.0.0.1")
}

func SharePhoneKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact("📱 Поделиться номером"),
		),
	)
}

func RemoveKeyboard() tgbotapi.ReplyKeyboardRemove {
	return tgbotapi.NewRemoveKeyboard(true)
}

// SharePhoneInlineMarkup — inline-кнопка «Поделиться номером» в приветственном сообщении.
// По нажатию бот отправит reply-клавиатуру с запросом контакта (так работает Telegram API).
func SharePhoneInlineMarkup() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📱 Поделиться номером", "share_phone"),
		),
	)
}

func OpenAppKeyboard(webAppURL string) tgbotapi.InlineKeyboardMarkup {
	// Используем URL-кнопку — в Telegram откроется WebView с Mini App.
	// Для нативной Mini App нужна кнопка web_app (если библиотека поддерживает).
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🎰 Открыть приложение", webAppURL),
		),
	)
}
