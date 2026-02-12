package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"

	"era_sporta_bot_ruletka/internal/api/middleware"
	"era_sporta_bot_ruletka/internal/notifier"
	"era_sporta_bot_ruletka/internal/service"

	"github.com/gin-gonic/gin"
)

const wheelSegments = 8
const disabledWheelPrizeName = "БЕЗЛИМИТ ПОСЕЩЕНИЙ НА 1 МЕСЯЦ"
const disabledWheelPrizeNameAlt = "1 МЕСЯЦ БЕСПЛАТНО"
const disabledWheelPrizeType = "free_month"

type RouletteHandler struct {
	rouletteSvc *service.RouletteService
	userSvc     *service.UserService
	adminNotify notifier.AdminNotifier
}

func NewRouletteHandler(rouletteSvc *service.RouletteService, userSvc *service.UserService, adminNotify notifier.AdminNotifier) *RouletteHandler {
	return &RouletteHandler{
		rouletteSvc: rouletteSvc,
		userSvc:     userSvc,
		adminNotify: adminNotify,
	}
}

func (h *RouletteHandler) Spin(c *gin.Context) {
	telegramUserID, ok := c.Get(middleware.TelegramUserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	tid := telegramUserID.(int64)

	ctx := c.Request.Context()
	user, err := h.userSvc.GetByTelegramID(ctx, tid)
	if err != nil || user == nil || user.Phone == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "phone required"})
		return
	}

	ipHash := hashIP(c.ClientIP())

	result, err := h.rouletteSvc.Spin(ctx, user.ID, ipHash)
	if err != nil {
		if errors.Is(err, service.ErrSpinLimitExceeded) {
			c.JSON(http.StatusConflict, gin.H{"error": "spin limit exceeded", "message": "Вы уже использовали свой спин"})
			return
		}
		log.Printf("[roulette] spin error for user_id=%d telegram_id=%d: %v", user.ID, tid, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Notify admin
	if h.adminNotify != nil {
		h.adminNotify.NotifySpin(ctx, user, result.Prize.Name)
	}

	stopSegment := h.resolveStopSegment(ctx, result.Prize.Name, result.ID)

	c.JSON(http.StatusOK, gin.H{
		"spin": gin.H{
			"id":           result.ID,
			"prize_id":     result.PrizeID,
			"prize_name":   result.Prize.Name,
			"prize_type":   result.Prize.Type,
			"value":        result.ResultValue,
			"created_at":   result.CreatedAt,
			"stop_segment": stopSegment,
		},
	})
}

func (h *RouletteHandler) Config(c *gin.Context) {
	ctx := c.Request.Context()
	prizes, err := h.rouletteSvc.GetConfig(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	type prizeDTO struct {
		ID     int     `json:"id"`
		Name   string  `json:"name"`
		Type   string  `json:"type"`
		Value  float64 `json:"value"`
		Weight int     `json:"weight"`
	}

	list := make([]prizeDTO, len(prizes))
	for i, p := range prizes {
		list[i] = prizeDTO{ID: p.ID, Name: p.Name, Type: p.Type, Value: p.Value, Weight: p.ProbabilityWeight}
	}
	c.JSON(http.StatusOK, gin.H{"prizes": list})
}

func (h *RouletteHandler) History(c *gin.Context) {
	telegramUserID, ok := c.Get(middleware.TelegramUserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	tid := telegramUserID.(int64)

	ctx := c.Request.Context()
	user, err := h.userSvc.GetByTelegramID(ctx, tid)
	if err != nil || user == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "phone required"})
		return
	}

	history, err := h.rouletteSvc.GetHistory(ctx, user.ID, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	type spinDTO struct {
		ID        int64   `json:"id"`
		PrizeName string  `json:"prize_name"`
		PrizeType string  `json:"prize_type"`
		Value     float64 `json:"value"`
		CreatedAt string  `json:"created_at"`
	}
	list := make([]spinDTO, len(history))
	for i, s := range history {
		list[i] = spinDTO{
			ID:        s.ID,
			PrizeName: s.Prize.Name,
			PrizeType: s.Prize.Type,
			Value:     s.ResultValue,
			CreatedAt: s.CreatedAt.Format("2006-01-02 15:04"),
		}
	}
	c.JSON(http.StatusOK, gin.H{"history": list})
}

func hashIP(ip string) string {
	if ip == "" {
		return ""
	}
	// Normalize (e.g. strip port)
	if idx := strings.LastIndex(ip, ":"); idx > 0 && strings.Contains(ip, ".") {
		// IPv4 with port
		ip = ip[:idx]
	}
	h := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(h[:])
}

func (h *RouletteHandler) resolveStopSegment(ctx context.Context, prizeName string, spinID int64) int {
	prizes, err := h.rouletteSvc.GetConfig(ctx)
	if err != nil || len(prizes) == 0 {
		return 0
	}
	segments := make([]*struct {
		name string
		typ  string
	}, wheelSegments)
	for i := 0; i < wheelSegments; i++ {
		p := prizes[i%len(prizes)]
		segments[i] = &struct {
			name string
			typ  string
		}{
			name: strings.TrimSpace(p.Name),
			typ:  strings.TrimSpace(p.Type),
		}
	}

	normalizedTarget := strings.ToUpper(strings.TrimSpace(prizeName))
	var preferred []int
	var fallback []int
	for idx, seg := range segments {
		if seg == nil {
			continue
		}
		if isBlockedWheelSegment(seg.name, seg.typ) {
			continue
		}
		fallback = append(fallback, idx)
		if strings.EqualFold(strings.ToUpper(seg.name), normalizedTarget) {
			preferred = append(preferred, idx)
		}
	}
	if len(preferred) > 0 {
		return preferred[stableModulo(spinID, len(preferred))]
	}
	if len(fallback) > 0 {
		return fallback[stableModulo(spinID, len(fallback))]
	}
	return 0
}

func isBlockedWheelSegment(name, prizeType string) bool {
	n := strings.ToUpper(strings.TrimSpace(name))
	t := strings.ToLower(strings.TrimSpace(prizeType))
	return n == disabledWheelPrizeName || n == disabledWheelPrizeNameAlt || t == disabledWheelPrizeType
}

func stableModulo(val int64, size int) int {
	if size <= 0 {
		return 0
	}
	n := val
	if n < 0 {
		n = -n
	}
	return int(n % int64(size))
}
