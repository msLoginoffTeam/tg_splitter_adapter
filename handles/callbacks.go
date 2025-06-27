package handles

//
//import (
//	"fmt"
//	"log"
//
//	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
//	client "github.com/msLoginoffTeam/tg_splitter_adapter/swagger"
//)
//
//func HandleCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery,
//	userStates map[int64]string, apiClient *client.ClientWithResponses) {
//
//	data := callbackQuery.Data
//	chatID := callbackQuery.Message.Chat.ID
//	userID := callbackQuery.From.ID
//
//	// Логируем callback для отладки
//	log.Printf("Callback from %d: %s", userID, data)
//
//	// Разбираем callback данные (ваша реализация может отличаться)
//	action, params, err := parseCallbackData(data)
//	if err != nil {
//		log.Printf("Error parsing callback data: %v", err)
//		return
//	}
//
//	// Обрабатываем разные типы действий
//	switch action {
//	case "group_select":
//		handleGroupSelect(bot, chatID, userID, params, userStates, apiClient)
//
//	case "expense_group_select":
//		handleExpenseGroupSelect(bot, chatID, userID, params, userStates)
//
//	case "add_expense":
//		handleAddExpense(bot, chatID, userID, params, userStates)
//
//	//case "edit_expense":
//	//	handleEditExpense(bot, chatID, userID, params, userStates)
//	//
//	//case "start_chat":
//	//	handleStartChat(bot, chatID, userID, params)
//
//	default:
//		log.Printf("Unknown callback action: %s", action)
//	}
//
//	// Отправляем подтверждение обработки callback
//	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
//	if _, err := bot.Request(callback); err != nil {
//		log.Printf("Callback confirmation error: %v", err)
//	}
//}
//
//// Вспомогательные функции для обработки конкретных действий
//
//func handleGroupSelect(bot *tgbotapi.BotAPI, chatID int64, userID int64,
//	params map[string]string, userStates map[int64]string,
//	apiClient *client.ClientWithResponses) {
//
//	groupID := params["group_id"]
//	groupName := params["group_name"]
//
//	// Здесь можно добавить запрос к API для получения информации о группе
//	// response, err := apiClient.GetGroupDetailsWithResponse(context.Background(), groupID)
//
//	msgText := fmt.Sprintf("Информация о группе %s (ID: %s):\n\nУчастники: ...\nБаланс: ...",
//		groupName, groupID)
//
//	// Создаем кнопки для действий с группой
//	keyboard := tgbotapi.NewInlineKeyboardMarkup(
//		tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonData("Начать общение",
//				createCallbackData("start_chat", map[string]string{
//					"group_id":   groupID,
//					"group_name": groupName,
//				})),
//			tgbotapi.NewInlineKeyboardButtonData("Показать траты",
//				createCallbackData("show_expenses", map[string]string{
//					"group_id": groupID,
//				})),
//		),
//	)
//
//	msg := tgbotapi.NewMessage(chatID, msgText)
//	msg.ReplyMarkup = keyboard
//	bot.Send(msg)
//}
//
//func handleExpenseGroupSelect(bot *tgbotapi.BotAPI, chatID int64, userID int64,
//	params map[string]string, userStates map[int64]string) {
//
//	groupID := params["group_id"]
//	groupName := params["group_name"]
//
//	// Устанавливаем состояние пользователя
//	userStates[userID] = "waiting_expense_action:" + groupID
//
//	// Создаем меню для работы с тратами
//	keyboard := tgbotapi.NewInlineKeyboardMarkup(
//		tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonData("Добавить трату",
//				createCallbackData("add_expense", map[string]string{
//					"group_id":   groupID,
//					"group_name": groupName,
//				})),
//			tgbotapi.NewInlineKeyboardButtonData("Изменить трату",
//				createCallbackData("edit_expense", map[string]string{
//					"group_id": groupID,
//				})),
//		),
//		tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonData("Показать все траты",
//				createCallbackData("show_expenses", map[string]string{
//					"group_id": groupID,
//				})),
//		),
//	)
//
//	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Выберите действие для группы %s:", groupName))
//	msg.ReplyMarkup = keyboard
//	bot.Send(msg)
//}
//
//func handleAddExpense(bot *tgbotapi.BotAPI, chatID int64, userID int64,
//	params map[string]string, userStates map[int64]string) {
//
//	groupID := params["group_id"]
//	userStates[userID] = "waiting_expense_amount:" + groupID
//
//	msg := tgbotapi.NewMessage(chatID, "Введите сумму траты:")
//	bot.Send(msg)
//}
//
//// Другие обработчики...
//
//// Функции для работы с callback данными
//
//func parseCallbackData(data string) (string, map[string]string, error) {
//	// Реализуйте парсинг вашего формата callback данных
//	// Например: "action:add_expense;group_id:123;group_name:Test"
//	params := make(map[string]string)
//	// ... логика парсинга ...
//	return params["action"], params, nil
//}
//
//func СreateCallbackData(action string, params map[string]string) string {
//	// Реализуйте создание строки callback данных
//	// Например: "action:add_expense;group_id:123;group_name:Test"
//	data := "action:" + action
//	for k, v := range params {
//		data += ";" + k + ":" + v
//	}
//	return data
//}
//
