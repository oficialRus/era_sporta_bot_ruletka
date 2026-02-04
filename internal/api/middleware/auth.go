package middleware

import (
	"net/http"
	"strings"

	"era_sporta_bot_ruletka/internal/telegram"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	botToken string
}

func NewAuthMiddleware(botToken string) *AuthMiddleware {
	return &AuthMiddleware{botToken: botToken}
}

type authKey struct{}

// TelegramUserID is the key for storing telegram user id in context
const TelegramUserIDKey = "telegram_user_id"

func (m *AuthMiddleware) InitDataAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		initData := c.GetHeader("X-Telegram-Init-Data")
		if initData == "" {
			initData = c.Query("initData")
		}
		if initData == "" {
			var body struct {
				InitData string `json:"initData"`
			}
			if err := c.ShouldBindJSON(&body); err == nil && body.InitData != "" {
				initData = body.InitData
			}
		}
		// Also check Authorization: Bearer <initData> for compatibility
		if initData == "" {
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				initData = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if initData == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "initData required"})
			c.Abort()
			return
		}

		result, err := telegram.ValidateInitData(initData, m.botToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "validation failed"})
			c.Abort()
			return
		}
		if !result.Valid || result.User == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid initData"})
			c.Abort()
			return
		}

		c.Set(TelegramUserIDKey, result.User.ID)
		c.Set("init_data_user", result.User)
		c.Next()
	}
}
