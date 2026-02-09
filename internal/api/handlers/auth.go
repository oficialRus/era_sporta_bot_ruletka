package handlers

import (
	"net/http"

	"era_sporta_bot_ruletka/internal/domain"
	"era_sporta_bot_ruletka/internal/service"
	"era_sporta_bot_ruletka/internal/telegram"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userSvc   *service.UserService
	botToken  string
	spinLimit int
}

func NewAuthHandler(userSvc *service.UserService, botToken string, spinLimit int) *AuthHandler {
	return &AuthHandler{userSvc: userSvc, botToken: botToken, spinLimit: spinLimit}
}

type AuthRequest struct {
	InitData string `json:"initData"`
}

type AuthResponse struct {
	User  *UserDTO      `json:"user"`
	State *UserStateDTO `json:"state"`
}

type UserDTO struct {
	ID             int64  `json:"id"`
	TelegramUserID int64  `json:"telegram_user_id"`
	Phone          string `json:"phone"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Username       string `json:"username"`
}

type UserStateDTO struct {
	SpinAvailable bool `json:"spin_available"`
	SpinsUsed     int  `json:"spins_used"`
	SpinLimit     int  `json:"spin_limit"`
}

func (h *AuthHandler) Auth(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.InitData == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "initData required"})
		return
	}

	result, err := telegram.ValidateInitData(req.InitData, h.botToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "validation failed"})
		return
	}
	if !result.Valid || result.User == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid initData"})
		return
	}

	ctx := c.Request.Context()
	user, err := h.userSvc.GetByTelegramID(ctx, result.User.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if user == nil || user.Phone == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "phone required", "message": "Сначала поделитесь номером телефона в боте"})
		return
	}

	state, err := h.userSvc.GetUserState(ctx, user, h.spinLimit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		User:  toUserDTO(user),
		State: toUserStateDTO(state),
	})
}

func toUserDTO(u *domain.User) *UserDTO {
	if u == nil {
		return nil
	}
	return &UserDTO{
		ID:             u.ID,
		TelegramUserID: u.TelegramUserID,
		Phone:          u.Phone,
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		Username:       u.Username,
	}
}

func toUserStateDTO(s *service.UserState) *UserStateDTO {
	if s == nil {
		return nil
	}
	return &UserStateDTO{
		SpinAvailable: s.SpinAvailable,
		SpinsUsed:     s.SpinsUsed,
		SpinLimit:     s.SpinLimit,
	}
}
