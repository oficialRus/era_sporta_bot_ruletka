package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

func OpenAppKeyboard(webAppURL string) tgbotapi.InlineKeyboardMarkup {
	// Используем URL-кнопку — в Telegram откроется WebView с Mini App.
	// Для нативной Mini App нужна кнопка web_app (если библиотека поддерживает).
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🎰 Открыть приложение", webAppURL),
		),
	)
}
