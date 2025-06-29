package handles

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/deepmap/oapi-codegen/pkg/types"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	tgutils "github.com/msLoginoffTeam/tg_splitter_adapter/handles/tg_utils"
	"github.com/msLoginoffTeam/tg_splitter_adapter/swagger"
	client "github.com/msLoginoffTeam/tg_splitter_adapter/swagger"
)

func HandleCommand(update *tgbotapi.Update, bot *tgbotapi.BotAPI, api *client.ClientWithResponses, adapter *tgutils.CommandAdapter) {

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	msg.ReplyToMessageID = update.Message.MessageID
	if !update.Message.IsCommand() {
		return
	}
	switch update.Message.Command() {
	case "start":
		msg.Text = "Для начала работы с ботом зайдите к нему в личные сообщения"
	case "register":
		userId := update.Message.From.ID

		newName := update.Message.From.FirstName + update.Message.From.LastName
		reqBody := client.UserCreateRequestDto{
			TelegramId:  &userId,
			DisplayName: &newName,
		}
		resp, err := api.PostApiUsers(context.Background(), reqBody)
		if err != nil {
			msg.Text = "Не получилось создать группу "
			break
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
			msg.Text = "Вы уже зарегистрированы"
			break
		}

		var result uuid.UUID
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Errorf("failed to decode response: %w", err)
			msg.Text = "Не удалось расшифровать ответ"
			break
		}
		msg.Text = "Успешно зарегистрирован пользователь с ником " + newName

	case "getmygroups":
		msg.Text = "/start — запуск\n/ping — проверка\n/help — помощь"

	case "creategroup":
		chatId := update.Message.Chat.ID
		userId := update.Message.From.ID
		chatTitle := update.Message.Chat.Title
		reqBody := client.CreateGroupRequestDto{
			CreatedByTelegramId: &userId,
			TelegramChatId:      &chatId,
			Title:               &chatTitle,
		}
		resp, err := api.PostApiGroups(context.Background(), reqBody)
		if err != nil {
			msg.Text = "Не получилось создать группу "
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
		msg.Text = "Группа успешно создана. Id группы:\n" + "Название группы: " + *result.Title + "\n" + "Id группы: " + (*result.Id).String()

	case "addtogroup":
		_, args := adapter.ParseCommand(update.Message.Text)
		if len(args) < 1 {
			msg.Text = "Не введен id группы"
			break
		}

		groupUUID, err := stringToUUID(args[0])
		if err != nil {
			msg.Text = "Не получилось обработать id"
			break
		}

		resp, err := api.PostApiGroupsGroupIdUsers(
			context.Background(),
			groupUUID,
			swagger.AddGroupUserRequestDto{
				TelegramId: &update.Message.From.ID,
			},
		)
		if err != nil {
			msg.Text = "Не получилось достучаться до API"
			fmt.Errorf("API call failed: %w", err)
			break
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			msg.Text = "Ошибка при добавлении пользователя, возможно вам надо зарегистрироваться /register"
			fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
			break
		}

		msg.Text = "Пользователь успешно добавлен"
	case "getchatgroups":
		tgChatId := update.Message.Chat.ID
		params := client.GetApiGroupsParams{
			TelegramChatId: &tgChatId,
		}
		resp, err := api.GetApiGroups(context.Background(), &params)
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
		if len(result) == 0 {
			msg.Text = "Нет групп привязанных к чату"
			break
		}
		msg.Text = "Группы, привязанные к чату: \n"
		for i, group := range result {
			msg.Text += strconv.Itoa(i+1) + ": \n"
			msg.Text += "Id_группы: `" + group.Id.String() + "`\n"
			msg.Text += "Название группы: " + *group.Title + "\n"
		}
		msg.ParseMode = "MarkdownV2"
		msg.Text = escapeMarkdown(msg.Text)
	case "groupdetails":
		_, args := adapter.ParseCommand(update.Message.Text)
		if len(args) < 1 {
			msg.Text = "Не введен id группы"
			break
		}
		groupId, err := stringToUUID(args[0])
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
	default:
		if update.Message.Command() != "" {
			msg.Text = fmt.Sprintf("Неизвестная команда: %s", update.Message.Command())
		}

	}
	bot.Send(msg)

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
