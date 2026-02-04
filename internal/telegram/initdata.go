package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// InitDataUser represents user data from Telegram WebApp initData
type InitDataUser struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
	IsPremium    bool   `json:"is_premium"`
}

// ValidateResult contains validation result and extracted user
type ValidateResult struct {
	Valid bool
	User  *InitDataUser
}

const maxInitDataAge = 24 * time.Hour

// ValidateInitData validates Telegram WebApp initData per official docs
// https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app
func ValidateInitData(initData string, botToken string) (*ValidateResult, error) {
	if initData == "" || botToken == "" {
		return &ValidateResult{Valid: false}, nil
	}

	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("parse initData: %w", err)
	}

	hash := values.Get("hash")
	if hash == "" {
		return &ValidateResult{Valid: false}, nil
	}
	values.Del("hash")

	// Build data_check_string: sorted key=value pairs joined by \n
	var keys []string
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+values.Get(k))
	}
	dataCheckString := strings.Join(parts, "\n")

	// secret_key = HMAC-SHA256("WebAppData", bot_token)
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))
	secretKeyHash := secretKey.Sum(nil)

	// hash = HMAC-SHA256(data_check_string, secret_key)
	h := hmac.New(sha256.New, secretKeyHash)
	h.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(h.Sum(nil))

	if hash != expectedHash {
		return &ValidateResult{Valid: false}, nil
	}

	// Check auth_date (not older than 24h)
	authDateStr := values.Get("auth_date")
	if authDateStr == "" {
		return &ValidateResult{Valid: false}, nil
	}
	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return &ValidateResult{Valid: false}, nil
	}
	if time.Since(time.Unix(authDate, 0)) > maxInitDataAge {
		return &ValidateResult{Valid: false}, nil
	}

	// Parse user
	userStr := values.Get("user")
	if userStr == "" {
		return &ValidateResult{Valid: true, User: nil}, nil
	}

	var user InitDataUser
	if err := json.Unmarshal([]byte(userStr), &user); err != nil {
		return &ValidateResult{Valid: false}, nil
	}

	return &ValidateResult{Valid: true, User: &user}, nil
}
