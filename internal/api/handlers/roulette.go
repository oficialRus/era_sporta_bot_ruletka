package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"era_sporta_bot_ruletka/internal/api/middleware"
	"era_sporta_bot_ruletka/internal/notifier"
	"era_sporta_bot_ruletka/internal/service"

	"github.com/gin-gonic/gin"
)

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Notify admin
	if h.adminNotify != nil {
		h.adminNotify.NotifySpin(ctx, user, result.Prize.Name)
	}

	c.JSON(http.StatusOK, gin.H{
		"spin": gin.H{
			"id":         result.ID,
			"prize_id":   result.PrizeID,
			"prize_name": result.Prize.Name,
			"prize_type": result.Prize.Type,
			"value":      result.ResultValue,
			"created_at": result.CreatedAt,
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
