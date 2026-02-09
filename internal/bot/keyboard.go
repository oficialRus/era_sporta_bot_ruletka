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

// SubscribeInlineMarkup — кнопки «Подписаться» + «Я подписался».
func SubscribeInlineMarkup(channelURL string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("📣 Подписаться", channelURL),
			tgbotapi.NewInlineKeyboardButtonData("✅ Я подписался", "check_subscribe"),
		),
	)
}

func OpenAppKeyboard(webAppURL string) map[string]interface{} {
	return map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{
					"text": "🎰 Открыть приложение",
					"web_app": map[string]string{
						"url": webAppURL,
					},
				},
			},
		},
	}
}
