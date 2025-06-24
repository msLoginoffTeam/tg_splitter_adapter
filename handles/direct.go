package handles

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	tgutils "github.com/msLoginoffTeam/tg_splitter_adapter/handles/tg_utils"
	client "github.com/msLoginoffTeam/tg_splitter_adapter/swagger"
)

func HandleDirectMessages(update *tgbotapi.Update, bot *tgbotapi.BotAPI, api *client.ClientWithResponses, adapter *tgutils.CommandAdapter, userStates map[int64]string) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	if update.Message.IsCommand() {
		switch update.Message.Command() {
		case "start":
			// Регистрация
			msg := tgbotapi.NewMessage(chatID, "Добро пожаловать! Вы зарегистрированы.")
			bot.Send(msg)

			// Отправка меню с кнопками
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
		case "Профиль":
			msg := tgbotapi.NewMessage(chatID, "Информация о вашем профиле:")
			bot.Send(msg)

		case "Группы пользователя":
			userStates[userID] = "waiting_group_selection"

			groups := []struct {
				ID   int
				Name string
			}{
				{1, "Группа 1"},
				{2, "Группа 2"},
				{3, "Группа 3"},
			}
			//логика для пагинации
			//// Создание кнопок для групп
			//var rows [][]tgbotapi.InlineKeyboardButton
			//for _, group := range groups {
			//	btn := tgbotapi.NewInlineKeyboardButtonData(group.Name,
			//		СreateCallbackData("group_select", map[string]string{
			//			"group_id":   strconv.Itoa(group.ID),
			//			"group_name": group.Name,
			//		}))
			//	rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
			//}
			//
			//keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
			//msg := tgbotapi.NewMessage(chatID, "Выберите группу:")
			//msg.ReplyMarkup = keyboard
			//bot.Send(msg)

		case "Траты в группе":
			userStates[userID] = "waiting_expense_group_selection"

			groups := []struct {
				ID   int
				Name string
			}{
				{1, "Группа 1"},
				{2, "Группа 2"},
				{3, "Группа 3"},
			}

			//логика для пагинации
			//var rows [][]tgbotapi.InlineKeyboardButton
			//for _, group := range groups {
			//	btn := tgbotapi.NewInlineKeyboardButtonData(group.Name,
			//		tgutils.CreateCallbackData("expense_group_select", map[string]string{
			//			"group_id":   strconv.Itoa(group.ID),
			//			"group_name": group.Name,
			//		}))
			//	rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
			//}
			//
			//keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
			//msg := tgbotapi.NewMessage(chatID, "Выберите группу для работы с тратами:")
			//msg.ReplyMarkup = keyboard
			//bot.Send(msg)
		default:
			//обработка состояний
			if state, ok := userStates[userID]; ok {
				switch state {
				case "waiting_group_selection":
					//выбор группы
				case "waiting_expense_group_selection":
					//выбор группы для выбора траты
				case "waiting_expense_selection":
					//выбор траты
				case "waiting_expense_amount":
					delete(userStates, userID)
				}
			} else {
				msg := tgbotapi.NewMessage(chatID, "Пожалуйста, используйте кнопки меню")
				bot.Send(msg)
			}
		}
	}
}
