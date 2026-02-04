package handlers

import (
	"net/http"

	"era_sporta_bot_ruletka/internal/api/middleware"
	"era_sporta_bot_ruletka/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userSvc   *service.UserService
	spinLimit int
}

func NewUserHandler(userSvc *service.UserService, spinLimit int) *UserHandler {
	return &UserHandler{userSvc: userSvc, spinLimit: spinLimit}
}

func (h *UserHandler) Me(c *gin.Context) {
	telegramUserID, ok := c.Get(middleware.TelegramUserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	tid := telegramUserID.(int64)

	ctx := c.Request.Context()
	user, err := h.userSvc.GetByTelegramID(ctx, tid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if user == nil || user.Phone == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "phone required"})
		return
	}

	state, err := h.userSvc.GetUserState(ctx, user, h.spinLimit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  toUserDTO(user),
		"state": toUserStateDTO(state),
	})
}

func (h *UserHandler) State(c *gin.Context) {
	// Same as Me, returns just state for Mini App
	h.Me(c)
}
