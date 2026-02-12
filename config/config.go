package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	BotToken            string
	AdminTelegramChatID int64
	TelegramChannelID   int64
	TelegramChannelURL  string
	WebAppURL           string
	APIPort             int
	DatabaseURL         string
	RedisURL            string
	RouletteSpinLimit   int
	RouletteLockTTLSec  int
}

func Load() (*Config, error) {
	c := &Config{
		BotToken:            getEnv("BOT_TOKEN", ""),
		AdminTelegramChatID: getEnvInt64("ADMIN_TELEGRAM_CHAT_ID", -5197400174),
		TelegramChannelID:   getEnvInt64("TELEGRAM_CHANNEL_ID", 0),
		TelegramChannelURL:  getEnv("TELEGRAM_CHANNEL_URL", ""),
		WebAppURL:           getEnv("WEBAPP_URL", "https://your-domain.com/webapp"),
		APIPort:             getEnvInt("API_PORT", 8080),
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		RedisURL:            getEnv("REDIS_URL", ""),
		RouletteSpinLimit:   getEnvInt("ROULETTE_SPIN_LIMIT_PER_USER", 1),
		RouletteLockTTLSec:  getEnvInt("ROULETTE_LOCK_TTL_SEC", 10),
	}
	return c, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return strings.TrimSpace(v)
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvInt64(key string, defaultVal int64) int64 {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil {
			return i
		}
	}
	return defaultVal
}
