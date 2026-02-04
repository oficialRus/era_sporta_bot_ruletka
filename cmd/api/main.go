package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"era_sporta_bot_ruletka/config"
	"era_sporta_bot_ruletka/internal/api"
	"era_sporta_bot_ruletka/internal/api/handlers"
	"era_sporta_bot_ruletka/internal/bot"
	"era_sporta_bot_ruletka/internal/notifier"
	"era_sporta_bot_ruletka/internal/db"
	"era_sporta_bot_ruletka/internal/repository"
	"era_sporta_bot_ruletka/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.BotToken == "" {
		log.Fatal("BOT_TOKEN is required")
	}
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	// Bot for admin notifications
	var adminNotify notifier.AdminNotifier
	if tgBot, err := tgbotapi.NewBotAPI(cfg.BotToken); err == nil {
		adminNotify = bot.NewAdminNotifierAdapter(bot.NewNotifier(tgBot, cfg.AdminTelegramChatID))
	} else {
		log.Printf("Warning: could not init bot for admin notifications: %v", err)
	}

	userRepo := repository.NewUserRepository(pool)
	prizeRepo := repository.NewPrizeRepository(pool)
	spinRepo := repository.NewSpinRepository(pool)

	userSvc := service.NewUserService(userRepo, spinRepo)
	rouletteSvc := service.NewRouletteService(pool, prizeRepo, spinRepo, userRepo, cfg.RouletteSpinLimit)

	authHandler := handlers.NewAuthHandler(userSvc, cfg.BotToken, cfg.RouletteSpinLimit)
	userHandler := handlers.NewUserHandler(userSvc, cfg.RouletteSpinLimit)
	rouletteHandler := handlers.NewRouletteHandler(rouletteSvc, userSvc, adminNotify)

	router := api.NewRouter(authHandler, userHandler, rouletteHandler, cfg.BotToken)

	app := gin.Default()
	router.Setup(app)

	addr := fmt.Sprintf(":%d", cfg.APIPort)
	srv := &http.Server{Addr: addr, Handler: app}

	go func() {
		log.Printf("API listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("API error: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	shutdownCtx := context.Background()
	return srv.Shutdown(shutdownCtx)
}
