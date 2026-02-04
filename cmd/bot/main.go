package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"era_sporta_bot_ruletka/config"
	"era_sporta_bot_ruletka/internal/bot"
	"era_sporta_bot_ruletka/internal/db"
	"era_sporta_bot_ruletka/internal/repository"
	"era_sporta_bot_ruletka/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

	tgBot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return err
	}
	log.Printf("Bot authorized as @%s", tgBot.Self.UserName)

	userRepo := repository.NewUserRepository(pool)
	spinRepo := repository.NewSpinRepository(pool)
	userSvc := service.NewUserService(userRepo, spinRepo)

	notifier := bot.NewNotifier(tgBot, cfg.AdminTelegramChatID)
	handler := bot.NewHandler(tgBot, userSvc, notifier, cfg.WebAppURL)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := tgBot.GetUpdatesChan(u)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		cancel()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-updates:
			handler.HandleUpdate(ctx, update)
		}
	}
}
