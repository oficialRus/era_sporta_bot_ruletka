package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Notifier struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

func NewNotifier(bot *tgbotapi.BotAPI, adminChatID int64) *Notifier {
	return &Notifier{bot: bot, chatID: adminChatID}
}

func (n *Notifier) Notify(ctx context.Context, text string) {
	if n.chatID == 0 {
		return
	}

	candidates := candidateChatIDs(n.chatID)
	for idx, chatID := range candidates {
		msg := tgbotapi.NewMessage(chatID, text)
		if _, err := n.bot.Send(msg); err != nil {
			if idx == len(candidates)-1 || !isRetryableChatIDError(err) {
				log.Printf("[notify] Send to admin error (chat_id=%d): %v", chatID, err)
				return
			}
			log.Printf("[notify] Retrying notification with fallback chat_id after error (chat_id=%d): %v", chatID, err)
			continue
		}
		if chatID != n.chatID {
			log.Printf("[notify] Delivered via fallback chat_id=%d (configured=%d)", chatID, n.chatID)
		}
		return
	}
}

func (n *Notifier) NotifyWithTime(ctx context.Context, text string) {
	n.Notify(ctx, text+"\nВремя: "+time.Now().Format("02.01.2006 15:04"))
}

func candidateChatIDs(chatID int64) []int64 {
	ids := []int64{chatID}
	// Some Telegram groups are returned in supergroup format (-100XXXXXXXXXX).
	if chatID < 0 && chatID > -1000000000000 {
		absID := -chatID
		if supergroupID, err := strconv.ParseInt(fmt.Sprintf("-100%d", absID), 10, 64); err == nil {
			ids = append(ids, supergroupID)
		}
	}
	return ids
}

func isRetryableChatIDError(err error) bool {
	if err == nil {
		return false
	}
	e := strings.ToLower(err.Error())
	return strings.Contains(e, "chat not found") ||
		strings.Contains(e, "group chat was upgraded to a supergroup chat") ||
		strings.Contains(e, "chat was upgraded to a supergroup")
}
