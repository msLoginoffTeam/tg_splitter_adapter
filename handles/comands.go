package handles

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleCommand(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {

	baseUrl := os.Getenv("BACKEND_URL")

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	msg.ReplyToMessageID = update.Message.MessageID
	switch update.Message.Command() {
	case "start":
		chatId := strconv.Itoa(int(update.Message.Chat.ID))
		userId := strconv.Itoa(int(update.Message.From.ID))
		msg.Text = "Id чата: " + chatId + "\n" + "Id пользователя: " + userId
	case "ping":
		msg.Text = "pong"
	case "help":
		msg.Text = "/start — запуск\n/ping — проверка\n/help — помощь"
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
