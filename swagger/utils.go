package swagger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

func GetUserUUIDbyid(api *ClientWithResponses, userID int64, msg tgbotapi.MessageConfig) *uuid.UUID {
	params := GetApiUsersFindParams{
		Nickname:       nil,
		UserTelegramId: &userID,
	}
	resp, err := api.GetApiUsersFind(context.Background(), &params)
	if err != nil {
		msg.Text = "API недоступно"
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		msg.Text = "Не удалось получить ответ от сервера"
		return nil
	}

	var result GroupResponseDto
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		msg.Text = "Не удалось расшифровать ответ"
		return nil
	}

	return result.Id

}

func GetGroupByUseridUtil(api *ClientWithResponses, userID int64, msg tgbotapi.MessageConfig) []GroupOverviewResponseDto {
	reqParams := GetApiGroupsMyParams{
		UserTelegramId: &userID,
	}
	resp, err := api.GetApiGroupsMy(context.Background(), &reqParams)
	if err != nil {
		msg.Text = "API недоступно"
		return make([]GroupOverviewResponseDto, 0)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
		msg.Text = "Не удалось получить ответ от сервера"
		return make([]GroupOverviewResponseDto, 0)
	}

	var result []GroupOverviewResponseDto
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Errorf("failed to decode response: %w", err)
		msg.Text = "Не удалось расшифровать ответ"
		return make([]GroupOverviewResponseDto, 0)
	}
	return result
}
func GetAllExpensesByGroupUtil(api *ClientWithResponses, userID int64, msg tgbotapi.MessageConfig, userChoiceState map[int64]uuid.UUID) []ExpenseResponseDto {
	resp, err := api.GetApiExpensesGroupGroupId(context.Background(), userChoiceState[userID], nil)
	if err != nil {
		msg.Text = "API недоступно"
		return make([]ExpenseResponseDto, 0)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
		msg.Text = "Не удалось получить ответ от сервера"
		return make([]ExpenseResponseDto, 0)
	}

	var result []ExpenseResponseDto
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Errorf("failed to decode response: %w", err)
		msg.Text = "Не удалось расшифровать ответ"
		return make([]ExpenseResponseDto, 0)
	}
	return result
}
func GetExpenseByIdUtil(api *ClientWithResponses, userID int64, expenseId uuid.UUID, msg tgbotapi.MessageConfig, userChoiceState map[int64]uuid.UUID) ExpenseResponseDto {

	idGroup := types.UUID(userChoiceState[userID])
	paramsExpense := GetApiExpensesExpenseIdParams{
		GroupId: &idGroup,
	}
	resp, err := api.GetApiExpensesExpenseId(context.Background(), expenseId, &paramsExpense)
	if err != nil {
		msg.Text = "API недоступно"
		return ExpenseResponseDto{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
		msg.Text = "Не удалось получить ответ от сервера"
		return ExpenseResponseDto{}
	}
	var result ExpenseResponseDto
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Errorf("failed to decode response: %w", err)
		msg.Text = "Не удалось расшифровать ответ"
		return ExpenseResponseDto{}
	}
	return result
}

func GetPaymentsByGroupIdUtil(api *ClientWithResponses, groupID uuid.UUID, msg tgbotapi.MessageConfig) []PaymentResponseDto {

	resp, err := api.GetApiGroupsGroupIdPayments(context.Background(), groupID)
	if err != nil {
		msg.Text = "API недоступно"
		return make([]PaymentResponseDto, 0)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
		msg.Text = "Не удалось получить ответ от сервера"
		return make([]PaymentResponseDto, 0)
	}
	var result []PaymentResponseDto
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Errorf("failed to decode response: %w", err)
		msg.Text = "Не удалось расшифровать ответ"
		return make([]PaymentResponseDto, 0)
	}
	return result
}

func stringToUUID(s string) (types.UUID, error) {
	// Парсим строку как UUID
	parsedUUID, err := uuid.Parse(s)
	if err != nil {
		return types.UUID{}, fmt.Errorf("invalid UUID string: %w", err)
	}

	// Преобразуем в types.UUID
	return types.UUID(parsedUUID), nil
}
