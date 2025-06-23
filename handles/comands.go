package handles

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	tgutils "github.com/msLoginoffTeam/tg_splitter_adapter/handles/tg_utils"
	client "github.com/msLoginoffTeam/tg_splitter_adapter/swagger"
)

func HandleCommand(update *tgbotapi.Update, bot *tgbotapi.BotAPI, api *client.ClientWithResponses, adapter *tgutils.CommandAdapter) {

	baseUrl := os.Getenv("BACKEND_URL")

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	msg.ReplyToMessageID = update.Message.MessageID
	switch update.Message.Command() {
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
		_, err := api.PostApiGroups(context.Background(), reqBody)
		if err != nil {
			msg.Text = "Не получилось создать группу " + err.Error()
		} else {
			msg.Text = "Группа успешно создана: " + chatTitle
		}

	case "addtogroup":
		//adaptergetGroupsParams := client.GetApiGroupsParams{
		//adapter	UserTelegramId: &update.Message.From.ID,
		//adapter}

	case "users":
		resp, err := http.Get(baseUrl + "/api/users")
		if err != nil {
			msg.Text = "Ошибка при запросе к серверу: " + err.Error()
			break
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			msg.Text = "Ошибка при чтении ответа: " + err.Error()
			break
		}

		if resp.StatusCode != http.StatusOK {
			msg.Text = fmt.Sprintf("Сервер вернул ошибку: %d\n%s", resp.StatusCode, string(body))
			break
		}

		msg.Text = string(body)

	default:
		msg.Text = fmt.Sprintf("Неизвестная команда: %s", update.Message.Command())
	}
	bot.Send(msg)

}
