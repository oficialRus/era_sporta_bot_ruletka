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
	text := fmt.Sprintf("🎰 Новый спин!\nНомер: %s\nЧто выиграл: %s", user.Phone, prizeName)
	a.notifier.NotifyWithTime(ctx, text)
}
