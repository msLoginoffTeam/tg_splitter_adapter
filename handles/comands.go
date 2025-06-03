package handles

import (
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Обработка команд
func HandleCommand(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	msg.ReplyToMessageID = update.Message.MessageID
	switch update.Message.Command() {
	case "start":
		chat_id := strconv.Itoa(int(update.Message.Chat.ID))
		user_id := strconv.Itoa(int(update.Message.From.ID))
		msg.Text = "Id чата: " + chat_id + "\n" + "Id пользователя: " + user_id
	case "ping":
		msg.Text = "pong"
	case "help":
		msg.Text = "/start — запуск\n/ping — проверка\n/help — помощь"
	default:
		msg.Text = fmt.Sprintf("Неизвестная команда: %s", update.Message.Command())
	}
	bot.Send(msg)
}
