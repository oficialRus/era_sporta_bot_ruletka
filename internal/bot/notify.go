package bot

import (
	"context"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Notifier struct {
	bot   *tgbotapi.BotAPI
	chatID int64
}

func NewNotifier(bot *tgbotapi.BotAPI, adminChatID int64) *Notifier {
	return &Notifier{bot: bot, chatID: adminChatID}
}

func (n *Notifier) Notify(ctx context.Context, text string) {
	if n.chatID == 0 {
		return
	}
	msg := tgbotapi.NewMessage(n.chatID, text)
	if _, err := n.bot.Send(msg); err != nil {
		log.Printf("[notify] Send to admin error: %v", err)
	}
}

func (n *Notifier) NotifyWithTime(ctx context.Context, text string) {
	n.Notify(ctx, text+"\nВремя: "+time.Now().Format("02.01.2006 15:04"))
}
