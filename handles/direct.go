package handles

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	tgutils "github.com/msLoginoffTeam/tg_splitter_adapter/handles/tg_utils"
	"github.com/msLoginoffTeam/tg_splitter_adapter/swagger"
	client "github.com/msLoginoffTeam/tg_splitter_adapter/swagger"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func HandleDirectMessages(update *tgbotapi.Update, bot *tgbotapi.BotAPI, api *client.ClientWithResponses, adapter *tgutils.CommandAdapter, userStates map[int64]string, userChoiceState map[int64]uuid.UUID, userChoiceTitleState map[int64]string, userExpenceCreated map[int64]uuid.UUID) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	if update.Message.IsCommand() {
		switch update.Message.Command() {
		case "start":
			// регистрация(добавить)
			msg := tgbotapi.NewMessage(chatID, "Добро пожаловать! Вы зарегистрированы.")
			bot.Send(msg)

			// мейн менюшка с кнопками
			mainMenu := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Профиль"),
					tgbotapi.NewKeyboardButton("Группы пользователя"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Траты в группе"),
				),
			)
			msg = tgbotapi.NewMessage(chatID, "Выберите действие:")
			msg.ReplyMarkup = mainMenu
			bot.Send(msg)

		default:
			msg := tgbotapi.NewMessage(chatID, "Неизвестная команда")
			bot.Send(msg)
		}
	} else {
		// обработка кнопок
		switch update.Message.Text {
		case "Вернуться в меню":
			msg := tgbotapi.NewMessage(chatID, "Главное меню")
			mainMenu := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Профиль"),
					tgbotapi.NewKeyboardButton("Группы пользователя"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Траты в группе"),
				),
			)
			msg = tgbotapi.NewMessage(chatID, "Выберите действие:")
			msg.ReplyMarkup = mainMenu
			bot.Send(msg)
		case "Профиль":
			msg := tgbotapi.NewMessage(chatID, "Информация о вашем профиле:")
			bot.Send(msg)

		case "Группы пользователя":
			userStates[userID] = "waiting_group_selection"

			msg := tgbotapi.NewMessage(chatID, "")

			reqParams := client.GetApiGroupsParams{
				UserTelegramId: &userID,
			}
			resp, err := api.GetApiGroups(context.Background(), &reqParams)
			if err != nil {
				msg.Text = "API недоступно"
				bot.Send(msg)
				break
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				body, _ := io.ReadAll(resp.Body)
				fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
				msg.Text = "Не удалось получить ответ от сервера"
				bot.Send(msg)
				break
			}

			var result []swagger.GroupOverviewResponseDto
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				fmt.Errorf("failed to decode response: %w", err)
				msg.Text = "Не удалось расшифровать ответ"
				bot.Send(msg)
				break
			}

			msg.Text = "Выберите группу для получения подробной информации:\n"

			for i, group := range result {
				msg.Text += strconv.Itoa(i+1) + ". Название: " + *group.Title + " \nId группы: " + group.Id.String() + "\n"
			}

			bot.Send(msg)
		case "Траты в группе":
			userStates[userID] = "waiting_expense_group_selection"
			msg := tgbotapi.NewMessage(chatID, "")

			reqParams := client.GetApiGroupsParams{
				UserTelegramId: &userID,
			}
			resp, err := api.GetApiGroups(context.Background(), &reqParams)
			if err != nil {
				msg.Text = "API недоступно"
				bot.Send(msg)
				break
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				body, _ := io.ReadAll(resp.Body)
				fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
				msg.Text = "Не удалось получить ответ от сервера"
				bot.Send(msg)
				break
			}

			var result []swagger.GroupOverviewResponseDto
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				fmt.Errorf("failed to decode response: %w", err)
				msg.Text = "Не удалось расшифровать ответ"
				bot.Send(msg)
				break
			}

			msg.Text = "Выберите группу для работы с тратами:\n"

			for i, group := range result {
				msg.Text += strconv.Itoa(i+1) + ". Название: " + *group.Title + " \nId группы: " + group.Id.String() + "\n"
			}

			bot.Send(msg)
		case "Добавить трату":
			userStates[userID] = "waiting_expense_title"
			msg := tgbotapi.NewMessage(chatID, "Выберите название для работы с тратами")
			bot.Send(msg)

		case "Изменить трату":
		default:
			//обработка состояний
			if state, ok := userStates[userID]; ok {
				switch state {
				case "waiting_group_selection":
					msg := tgbotapi.NewMessage(chatID, "")
					groupId, err := stringToUUID(update.Message.Text)
					if err != nil {
						msg.Text = "Не получилось обработать id"
						break
					}

					resp, err := api.GetApiGroupsGroupId(context.Background(), groupId)
					if err != nil {
						msg.Text = "API недоступно"
						break
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
						body, _ := io.ReadAll(resp.Body)
						fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
						msg.Text = "Не удалось получить ответ от сервера"
						break
					}

					var result swagger.GroupResponseDto
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						fmt.Errorf("failed to decode response: %w", err)
						msg.Text = "Не удалось расшифровать ответ"
						break
					}

					msg.Text = fmt.Sprintf("Название группы: %s \n Id группы: %s \n Пользователи: \n", *result.Title, *result.Id)

					for i, member := range *result.Users {
						msg.Text += strconv.Itoa(i+1) + ". " + "Id: " + member.Id.String() + "\n Имя в системе: " + *member.DisplayName + "\n"
					}
					bot.Send(msg)
				case "waiting_expense_group_selection":
					msg := tgbotapi.NewMessage(chatID, "")
					groupId, err := stringToUUID(update.Message.Text)
					if err != nil {
						msg.Text = "Не получилось обработать id"
						break
					}

					resp, err := api.GetApiGroupsGroupId(context.Background(), groupId)
					if err != nil {
						msg.Text = "API недоступно"
						break
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
						body, _ := io.ReadAll(resp.Body)
						fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
						msg.Text = "Не удалось получить ответ от сервера"
						break
					}

					var result swagger.GroupResponseDto
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						fmt.Errorf("failed to decode response: %w", err)
						msg.Text = "Не удалось расшифровать ответ"
						break
					}

					userChoiceState[userID] = groupId

					mainMenu := tgbotapi.NewReplyKeyboard(
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton("Добавить трату"),
							tgbotapi.NewKeyboardButton("Изменить трату"),
						),
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton("Вернуться в меню"),
						),
					)
					msg = tgbotapi.NewMessage(chatID, "Выберите действие с тратами:")
					msg.ReplyMarkup = mainMenu
					bot.Send(msg)
				case "waiting_expense_selection":
					//выбор траты
				case "waiting_expense_title":
					userChoiceTitleState[userID] = update.Message.Text
					userStates[userID] = "waiting_expense_amount"
					bot.Send(tgbotapi.NewMessage(chatID, "Введите общую сумму траты"))

				case "waiting_expense_amount":

					msg := tgbotapi.NewMessage(chatID, "")
					creatorId, err := swagger.GetUserUUIDbyid(api, userID, msg)
					if err != nil {
						bot.Send(msg)
						break
					}
					isDraft := false
					Title := userChoiceTitleState[userID]
					TotalAmount, err := strconv.Atoi(update.Message.Text)
					if err != nil {
						bot.Send(tgbotapi.NewMessage(chatID, "Неприавльно введена сумма, попробуйте заново"))
						break
					}
					TotalAmountFloat := float64(TotalAmount)
					SharesEmpty := make([]client.ExpenseShareCreateDto, 0)
					reqBody := client.CreateExpenseRequestDto{
						CreatedById: creatorId,
						IsDraft:     &isDraft,
						Title:       &Title,
						TotalAmount: &TotalAmountFloat,
						Shares:      &SharesEmpty,
					}
					resp, err := api.PostApiGroupsGroupIdExpenses(context.Background(), userChoiceState[userID], reqBody)
					if err != nil {
						msg.Text = "Не получилось создать трату"
						bot.Send(msg)
						break
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
						body, _ := io.ReadAll(resp.Body)
						fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
						msg.Text = "Не удалось получить ответ от сервера"
						bot.Send(msg)
						break
					}

					var result swagger.ExpenseResponseDto
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						fmt.Errorf("failed to decode response: %w", err)
						msg.Text = "Не удалось расшифровать ответ"
						bot.Send(msg)
						break
					}
					expenseId, err := uuid.Parse(result.Id.String())
					if err != nil {
						msg.Text = "Не удалось расшифровать ответ"
						bot.Send(msg)
						break
					}
					userExpenceCreated[userID] = expenseId
					userStates[userID] = "waiting_expense_users"

					msg.Text = "Введите через пробел сумму которую должен человек и id человека"
					bot.Send(msg)
				case "waiting_expense_users":
					msg := tgbotapi.NewMessage(chatID, "")
					args := strings.Fields(update.Message.Text)
					if len(args) != 2 {
						msg.Text = "Предоставлено не два аргумента"
						bot.Send(msg)
						break
					}

					amount, err := strconv.Atoi(args[0])
					if err != nil {
						msg.Text = "Сумма указана неверно, попробуйте заново"
						bot.Send(msg)
						break
					}
					userId, err := stringToOpenapiUUIDPtr(args[1])
					if err != nil {
						msg.Text = "Id указан неверно, попробуйте заново"
						bot.Send(msg)
						break
					}

					amountFloat := float64(amount)
					reqBody := client.ExpenseShareCreateDto{
						Amount: &amountFloat,
						UserId: userId,
					}
					resp, err := api.PostApiGroupsGroupIdExpensesExpenseIdParticipants(context.Background(), userChoiceState[userID], userExpenceCreated[userID], reqBody)
					if err != nil {
						msg.Text = "Не получилось создать трату"
						bot.Send(msg)
						break
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
						body, _ := io.ReadAll(resp.Body)
						fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
						msg.Text = "Не удалось получить ответ от сервера"
						bot.Send(msg)
						break
					}
					mainMenu := tgbotapi.NewReplyKeyboard(
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton("Продолжить"),
						),
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton("Вернуться в меню"),
						),
					)
					msg.Text = "Пользователь добавлен в трату, хотите продолжить добавление или закончить:"
					msg.ReplyMarkup = mainMenu
					bot.Send(msg)
				}

			} else {
				msg := tgbotapi.NewMessage(chatID, "Пожалуйста, используйте кнопки меню")
				bot.Send(msg)
			}
		}
	}
}

func stringToOpenapiUUIDPtr(uuidStr string) (*openapi_types.UUID, error) {
	if uuidStr == "" {
		return nil, nil
	}

	stdUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, err
	}

	openapiUUID := openapi_types.UUID(stdUUID)
	return &openapiUUID, nil
}
