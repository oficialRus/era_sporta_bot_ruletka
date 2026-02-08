package api

import (
	"era_sporta_bot_ruletka/internal/api/handlers"
	"era_sporta_bot_ruletka/internal/api/middleware"
	"era_sporta_bot_ruletka/internal/bot"

	"github.com/gin-gonic/gin"
)

type Router struct {
	authHandler    *handlers.AuthHandler
	userHandler    *handlers.UserHandler
	rouletteHandler *handlers.RouletteHandler
	authMiddleware *middleware.AuthMiddleware
}

func NewRouter(
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	rouletteHandler *handlers.RouletteHandler,
	botToken string,
) *Router {
	return &Router{
		authHandler:     authHandler,
		userHandler:     userHandler,
		rouletteHandler: rouletteHandler,
		authMiddleware:  middleware.NewAuthMiddleware(botToken),
	}
}

func (r *Router) Setup(app *gin.Engine) {
	app.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Telegram-Init-Data")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	api := app.Group("/api")
	{
		// Public
		api.POST("/auth", r.authHandler.Auth)
		api.GET("/roulette/config", r.rouletteHandler.Config)

		// Protected (require initData)
		protected := api.Group("")
		protected.Use(r.authMiddleware.InitDataAuth())
		{
			protected.GET("/user/me", r.userHandler.Me)
			protected.GET("/user/state", r.userHandler.State)
			protected.POST("/roulette/spin", r.rouletteHandler.Spin)
			protected.GET("/roulette/history", r.rouletteHandler.History)
		}
	}
}

// Dependencies holds all deps for API - used by cmd/api
type Dependencies struct {
	AuthHandler     *handlers.AuthHandler
	UserHandler     *handlers.UserHandler
	RouletteHandler *handlers.RouletteHandler
	BotHandler      *bot.Handler
}
