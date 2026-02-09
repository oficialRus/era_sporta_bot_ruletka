package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type chatMemberResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		Status string `json:"status"`
	} `json:"result"`
	Description string `json:"description"`
}

func IsUserMember(ctx context.Context, botToken string, chatID int64, userID int64) (bool, error) {
	if botToken == "" || chatID == 0 || userID == 0 {
		return false, fmt.Errorf("invalid telegram params")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getChatMember?chat_id=%d&user_id=%d", botToken, chatID, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var payload chatMemberResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return false, err
	}
	if !payload.Ok {
		return false, fmt.Errorf("telegram api error: %s", payload.Description)
	}

	switch payload.Result.Status {
	case "left", "kicked":
		return false, nil
	default:
		return true, nil
	}
}
