package swagger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

func GetUserUUIDbyid(api *ClientWithResponses, userID int64, msg tgbotapi.MessageConfig) (*uuid.UUID, error) {
	params := GetApiUsersFindParams{
		Nickname:       nil,
		UserTelegramId: &userID,
	}
	resp, err := api.GetApiUsersFind(context.Background(), &params)
	if err != nil {
		msg.Text = "API недоступно"
		return nil, fmt.Errorf("API недоступно")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		msg.Text = "Не удалось получить ответ от сервера"
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result GroupResponseDto
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Errorf("failed to decode response: %w", err)
		msg.Text = "Не удалось расшифровать ответ"
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Id, nil

}
