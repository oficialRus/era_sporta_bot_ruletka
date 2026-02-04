package bot

import (
	"context"
	"fmt"

	"era_sporta_bot_ruletka/internal/domain"
)

// AdminNotifierAdapter adapts Notifier to notifier.AdminNotifier interface
type AdminNotifierAdapter struct {
	notifier *Notifier
}

func NewAdminNotifierAdapter(notifier *Notifier) *AdminNotifierAdapter {
	return &AdminNotifierAdapter{notifier: notifier}
}

func (a *AdminNotifierAdapter) NotifySpin(ctx context.Context, user *domain.User, prizeName string) {
	if a.notifier == nil {
		return
	}
	text := fmt.Sprintf("🎰 Новый спин!\nПользователь: %s %s (@%s)\nТелефон: %s\nПриз: %s",
		user.FirstName, user.LastName, user.Username, user.Phone, prizeName)
	a.notifier.NotifyWithTime(ctx, text)
}
