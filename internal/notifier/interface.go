package notifier

import (
	"context"

	"era_sporta_bot_ruletka/internal/domain"
)

// AdminNotifier notifies admin about spin results
type AdminNotifier interface {
	NotifySpin(ctx context.Context, user *domain.User, prizeName string)
}
