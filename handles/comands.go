package handles

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	client "github.com/msLoginoffTeam/tg_splitter_adapter/swagger"
)

func HandleCommand(update *tgbotapi.Update, bot *tgbotapi.BotAPI, api *client.ClientWithResponses) {

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
		getGroupsParams := client.GetApiGroupsParams{
			UserTelegramId: &update.Message.From.ID,
		}
		groups, err := api.GetApiGroups(context.Background(), &getGroupsParams)
		if err != nil {
			log.Printf("Ошибка при получении групп: %v", err)
			msg.Text = "Не удалось загрузить список групп. Попробуйте позже."
			break
		}

		if groups == nil {
			log.Println("Ответ от сервера пустой")
			msg.Text = "Не удалось получить ответ от сервера"
			break
		}
		defer groups.Body.Close()

		if groups.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(groups.Body)
			log.Printf("Сервер вернул ошибку: %d, тело: %s", groups.StatusCode, string(body))
			msg.Text = "Сервер вернул ошибку"
			break
		}

		var groupList []client.GroupOverviewResponseDto
		if err := json.NewDecoder(groups.Body).Decode(&groupList); err != nil {
			log.Printf("Ошибка при декодировании JSON: %v", err)
			msg.Text = "Ошибка при обработке данных групп"
			break
		}

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
