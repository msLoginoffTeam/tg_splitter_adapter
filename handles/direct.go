package handles

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	tgutils "github.com/msLoginoffTeam/tg_splitter_adapter/handles/tg_utils"
	"github.com/msLoginoffTeam/tg_splitter_adapter/swagger"
	client "github.com/msLoginoffTeam/tg_splitter_adapter/swagger"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func HandleDirectMessages(update *tgbotapi.Update, bot *tgbotapi.BotAPI, api *client.ClientWithResponses, adapter *tgutils.CommandAdapter, userStates map[int64]string, userChoiceState map[int64]uuid.UUID, userChoiceTitleState map[int64]string, userExpenceCreated map[int64]uuid.UUID, userSysID map[int64]uuid.UUID, userPaymentID map[int64]uuid.UUID) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID
	mainMenu := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Профиль"),
			tgbotapi.NewKeyboardButton("Группы"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Справка о тратах"),
			tgbotapi.NewKeyboardButton("Платежи"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Траты в группе"),
		),
	)

	if update.Message.IsCommand() {
		switch update.Message.Command() {
		case "start":
			file, err := os.Open("C:/Projects/tg_splitter_adapter/introduce.txt")
			if err != nil {
				log.Fatal(err)
			}
			defer func() {
				if err = file.Close(); err != nil {
					log.Fatal(err)
				}
			}()

			b, err := io.ReadAll(file)
			if err != nil {
				log.Fatal(err)
			}

			text := string(b)
			msg := tgbotapi.NewMessage(chatID, text)
			bot.Send(msg)
		case "register":
			// регистрация при начале общения с ботом
			msg := tgbotapi.NewMessage(chatID, "Добро пожаловать! Вы зарегистрированы.")
			newName := update.Message.From.FirstName + update.Message.From.LastName
			reqBody := client.UserCreateRequestDto{
				TelegramId:  &userID,
				DisplayName: &newName,
			}
			resp, err := api.PostApiUsers(context.Background(), reqBody)
			if err != nil {
				msg.Text = "Сервер не отвечает"
				break
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				body, _ := io.ReadAll(resp.Body)
				fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
				msg = tgbotapi.NewMessage(chatID, "Выберите действие:")
				msg.ReplyMarkup = mainMenu
				bot.Send(msg)
				break
			}

			var result uuid.UUID
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				fmt.Errorf("failed to decode response: %w", err)
				msg = tgbotapi.NewMessage(chatID, "Выберите действие:")
				msg.ReplyMarkup = mainMenu
				bot.Send(msg)
				break
			}
			msg.Text = "Успешно зарегистрирован пользователь с ником " + newName + "\n Выберите действие из меню:"

			msg.ReplyMarkup = mainMenu
			bot.Send(msg)
		default:
			msg := tgbotapi.NewMessage(chatID, "Эта команда доступна только в чате")
			bot.Send(msg)
		}
	} else {
		// обработка кнопок
		switch update.Message.Text {
		case "Вернуться в меню":
			msg := tgbotapi.NewMessage(chatID, "Главное меню")
			msg = tgbotapi.NewMessage(chatID, "Выберите действие:")
			msg.ReplyMarkup = mainMenu
			bot.Send(msg)
		case "Переименоваться":
			userStates[userID] = "waiting_username_selection"
			msg := tgbotapi.NewMessage(chatID, "Напишите новое имя")
			bot.Send(msg)
		case "Профиль":
			msg := tgbotapi.NewMessage(chatID, "Напишите новое имя")
			emptyNick := ""
			reqParams := client.GetApiUsersFindParams{
				UserTelegramId: &userID,
				Nickname:       &emptyNick,
			}
			resp, err := api.GetApiUsersFind(context.Background(), &reqParams)
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

			var result client.UserResponseDto
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				fmt.Errorf("failed to decode response: %w", err)
				msg.Text = "Не удалось расшифровать ответ"
				bot.Send(msg)
				break
			}

			profMenu := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Переименоваться"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Вернуться в меню"),
				),
			)
			msg = tgbotapi.NewMessage(
				chatID,
				fmt.Sprintf(
					"Информация о вашем профиле:\n"+
						"Имя: %s\n"+
						"ID юзера: `%s`", // UUID в code-блоке
					*result.DisplayName,
					result.Id.String(), // UUID не нужно экранировать, так как он в code-блоке
				),
			)
			msg.ParseMode = "MarkdownV2"
			userSysID[userID] = *result.Id
			msg.ReplyMarkup = profMenu
			bot.Send(msg)

		case "Справка о тратах":
			summaryMenu := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Баланс"),
					tgbotapi.NewKeyboardButton("Трансферы"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Вернуться в меню"),
				),
			)
			msg := tgbotapi.NewMessage(chatID, "Выберите действие")
			msg.ReplyMarkup = summaryMenu
			bot.Send(msg)

		case "Баланс":
			msg := tgbotapi.NewMessage(chatID, "")

			result := client.GetGroupByUseridUtil(api, userID, msg)
			if msg.Text != "" {
				bot.Send(msg)
				break
			}
			if len(result) == 0 {
				msg.Text = "Группы не найдены - создайте!!"
				bot.Send(msg)
				break
			}

			msg.Text = "Выберите группу:\n"

			for i, group := range result {
				msg.Text += strconv.Itoa(i+1) + "\\. Название: " + *group.Title + " \nId группы: `" + group.Id.String() + "`\n"
			}
			msg.ParseMode = "MarkdownV2"
			userStates[userID] = "waiting_balance_group_selection"
			bot.Send(msg)
		case "Трансферы":
			userStates[userID] = "waiting_transfer_group_selection"
			msg := tgbotapi.NewMessage(chatID, "")

			result := client.GetGroupByUseridUtil(api, userID, msg)
			if msg.Text != "" {
				bot.Send(msg)
				break
			}
			if len(result) == 0 {
				msg.Text = "Группы не найдены - создайте!!"
				bot.Send(msg)
				break
			}

			msg.Text = "Выберите группу:\n"

			for i, group := range result {
				msg.Text += strconv.Itoa(i+1) + "\\. Название: " + *group.Title + " \nId группы: `" + group.Id.String() + "`\n"
			}
			msg.ParseMode = "MarkdownV2"

			bot.Send(msg)
		case "Платежи":
			userStates[userID] = "waiting_payments_group_selection"

			msg := tgbotapi.NewMessage(chatID, "")

			result := client.GetGroupByUseridUtil(api, userID, msg)
			if msg.Text != "" {
				bot.Send(msg)
				break
			}
			if len(result) == 0 {
				msg.Text = "Группы не найдены - создайте!!"
				bot.Send(msg)
				break
			}

			msg.Text = "Выберите группу:\n"

			for i, group := range result {
				msg.Text += strconv.Itoa(i+1) + "\\. Название: " + *group.Title + " \nId группы: `" + group.Id.String() + "`\n"
			}
			msg.ParseMode = "MarkdownV2"

			bot.Send(msg)

		case "Группы пользователя":
			userStates[userID] = "waiting_group_selection"

			msg := tgbotapi.NewMessage(chatID, "")

			result := client.GetGroupByUseridUtil(api, userID, msg)
			if msg.Text != "" {
				bot.Send(msg)
				break
			}
			if len(result) == 0 {
				msg.Text = "Группы не найдены - создайте!!"
				break
			}

			msg.Text = "Выберите группу:\n"

			for i, group := range result {
				msg.Text += strconv.Itoa(i+1) + "\\. Название: " + *group.Title + " \nId группы: `" + group.Id.String() + "`\n"
			}
			msg.ParseMode = "MarkdownV2"

			bot.Send(msg)
		case "Присоединиться к группе":
			userStates[userID] = "waiting_group_id_join"
			msg := tgbotapi.NewMessage(chatID, "Напишите id группы, к которой хотите присоединиться")
			bot.Send(msg)
		case "Создать группу":
			userStates[userID] = "waiting_group_name_create_selection"
			msg := tgbotapi.NewMessage(chatID, "Напишите название вашей новой группы")
			bot.Send(msg)
		case "Изменить группу":
			userStates[userID] = "waiting_group_edit_selection"
			msg := tgbotapi.NewMessage(chatID, "")

			result := client.GetGroupByUseridUtil(api, userID, msg)
			if msg.Text != "" {
				bot.Send(msg)
				break
			}
			if len(result) == 0 {
				msg.Text = "Группы не найдены - создайте!!"
				bot.Send(msg)
				break
			}

			msg.Text = "Выберите группу:\n"

			for i, group := range result {
				msg.Text += strconv.Itoa(i+1) + "\\. Название: " + *group.Title + " \nId группы: `" + group.Id.String() + "`\n"
			}
			msg.ParseMode = "MarkdownV2"

			bot.Send(msg)
		case "Удалить группу":
			userStates[userID] = "waiting_group_delete_selection"
			msg := tgbotapi.NewMessage(chatID, "")

			result := client.GetGroupByUseridUtil(api, userID, msg)
			if msg.Text != "" {
				bot.Send(msg)
				break
			}
			if len(result) == 0 {
				msg.Text = "Группы не найдены - создайте!!"
				break
			}

			msg.Text = "Выберите группу:\n"

			for i, group := range result {
				msg.Text += strconv.Itoa(i+1) + "\\. Название: " + *group.Title + " \nId группы: `" + group.Id.String() + "`\n"
			}
			msg.ParseMode = "MarkdownV2"

			bot.Send(msg)

		case "Изменить/Удалить группу":
			summaryMenu := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Изменить группу"),
					tgbotapi.NewKeyboardButton("Удалить группу"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Вернуться в меню"),
				),
			)
			msg := tgbotapi.NewMessage(chatID, "Выберите действие")
			msg.ReplyMarkup = summaryMenu
			bot.Send(msg)
		case "Группы":
			summaryMenu := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Группы пользователя"),
					tgbotapi.NewKeyboardButton("Создать группу"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Присоединиться к группе"),
					tgbotapi.NewKeyboardButton("Изменить/Удалить группу"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Вернуться в меню"),
				),
			)
			msg := tgbotapi.NewMessage(chatID, "Выберите действие")
			msg.ReplyMarkup = summaryMenu
			bot.Send(msg)
		case "Траты в группе":
			userStates[userID] = "waiting_expense_group_selection"
			msg := tgbotapi.NewMessage(chatID, "")

			result := client.GetGroupByUseridUtil(api, userID, msg)
			if msg.Text != "" {
				bot.Send(msg)
				break
			}

			if len(result) == 0 {
				msg.Text = "Группы не найдены - создайте!!"
				break
			}

			msg.Text = "Выберите группу:\n"

			for i, group := range result {
				msg.Text += strconv.Itoa(i+1) + "\\. Название: " + *group.Title + " \nId группы: `" + group.Id.String() + "`\n"
			}
			msg.ParseMode = "MarkdownV2"

			bot.Send(msg)
		case "Трата от своего лица":
			userStates[userID] = "waiting_expense_title"
			msg := tgbotapi.NewMessage(chatID, "Выберите название для работы с тратами")
			bot.Send(msg)
		case "Трата от 3-его лица":
			msg := tgbotapi.NewMessage(chatID, "")
			respGroups, err := api.GetApiGroupsGroupId(context.Background(), userChoiceState[userID])
			if err != nil {
				msg.Text = "API недоступно"
				bot.Send(msg)
				break
			}
			defer respGroups.Body.Close()

			if respGroups.StatusCode != http.StatusOK && respGroups.StatusCode != http.StatusCreated {
				body, _ := io.ReadAll(respGroups.Body)
				fmt.Errorf("unexpected status code: %d, body: %s", respGroups.StatusCode, string(body))
				msg.Text = "Не удалось получить ответ от сервера"
				bot.Send(msg)
				break
			}

			var resultGroups swagger.GroupResponseDto
			if err := json.NewDecoder(respGroups.Body).Decode(&resultGroups); err != nil {
				fmt.Errorf("failed to decode response: %w", err)
				msg.Text = "Не удалось расшифровать ответ"
				bot.Send(msg)
				break
			}

			msg.Text = fmt.Sprintf("Название группы: %s \n Id группы: `%s` \n Пользователи: \n", *resultGroups.Title, *resultGroups.Id)

			for i, member := range *resultGroups.Users {
				msg.Text += strconv.Itoa(i+1) + ". " + "Id: `" + member.Id.String() + "`\n Имя в системе: " + *member.DisplayName + "\n\n"
			}
			msg.Text += "Введите id пользователя от имени которого будет создаваться трата:"
			msg.ParseMode = "MarkdownV2"
			msg.Text = escapeMarkdown(msg.Text)
			userStates[userID] = "waiting_3rd_person_expense_user"
			bot.Send(msg)

		case "Добавить трату":
			summaryMenu := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Трата от своего лица"),
					tgbotapi.NewKeyboardButton("Трата от 3-его лица"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Вернуться в меню"),
				),
			)
			msg := tgbotapi.NewMessage(chatID, "Выберите действие")
			msg.ReplyMarkup = summaryMenu
			bot.Send(msg)

		case "Изменить трату":
			userStates[userID] = "waiting_expense_selection"
			msg := tgbotapi.NewMessage(chatID, "")

			result := client.GetAllExpensesByGroupUtil(api, userID, msg, userChoiceState)
			if msg.Text != "" {
				bot.Send(msg)
				break
			}
			if len(result) == 0 {
				msg.Text = "Траты не найдены - создайте!!"
				bot.Send(msg)
				break
			}
			msg.Text = "Список трат:\n"

			for i, expense := range result {
				msg.Text += strconv.Itoa(i+1) + ".\n"
				msg.Text += "Название траты: " + *expense.Title + "\n"
				msg.Text += "Общая сумма: " + strconv.Itoa(int(*expense.TotalAmount)) + "\n"
				msg.Text += "Id траты: `" + expense.Id.String() + "`\n\n"
			}
			msg.ParseMode = "MarkdownV2"
			msg.Text = escapeMarkdown(msg.Text)
			bot.Send(msg)
		case "Изменить название траты":
			userStates[userID] = "waiting_new_expense_title_users"
			msg := tgbotapi.NewMessage(chatID, "Введите новое название траты")
			bot.Send(msg)
		case "Изменить сумму траты":
			userStates[userID] = "waiting_new_expense_total_amount_users"
			msg := tgbotapi.NewMessage(chatID, "Введите новую сумму траты")
			bot.Send(msg)

		case "Добавить участников":
			msg := tgbotapi.NewMessage(chatID, "")
			respGroups, err := api.GetApiGroupsGroupId(context.Background(), userChoiceState[userID])
			if err != nil {
				msg.Text = "API недоступно"
				bot.Send(msg)
				break
			}
			defer respGroups.Body.Close()

			if respGroups.StatusCode != http.StatusOK && respGroups.StatusCode != http.StatusCreated {
				body, _ := io.ReadAll(respGroups.Body)
				fmt.Errorf("unexpected status code: %d, body: %s", respGroups.StatusCode, string(body))
				msg.Text = "Не удалось получить ответ от сервера"
				bot.Send(msg)
				break
			}

			var resultGroups swagger.GroupResponseDto
			if err := json.NewDecoder(respGroups.Body).Decode(&resultGroups); err != nil {
				fmt.Errorf("failed to decode response: %w", err)
				msg.Text = "Не удалось расшифровать ответ"
				bot.Send(msg)
				break
			}

			msg.Text = fmt.Sprintf("Название группы: %s \nПользователи: \n", *resultGroups.Title)

			for i, member := range *resultGroups.Users {
				msg.Text += strconv.Itoa(i+1) + ". " + "Id: `" + member.Id.String() + "`\n Имя в системе: " + *member.DisplayName + "\n\n"
			}
			msg.ParseMode = "MarkdownV2"
			msg.Text = escapeMarkdown(msg.Text)
			userStates[userID] = "waiting_expense_users"
			msg.Text += "Введите через пробел сумму которую должен человек и id человека"
			bot.Send(msg)

		case "Удалить трату":
			msg := tgbotapi.NewMessage(chatID, "")

			resp, err := api.DeleteApiExpensesExpenseId(context.Background(), userChoiceState[userID], nil)
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

			msg.Text = "Трата успешно удалена"
			bot.Send(msg)
		case "Создать оплату":
			msg := tgbotapi.NewMessage(chatID, "")
			expenseMenu := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Привязанный перевод"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Свободный перевод"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Вернуться в меню"),
				),
			)
			msg = tgbotapi.NewMessage(chatID, "Привязать перевод к трате или нет:")
			msg.ReplyMarkup = expenseMenu
			bot.Send(msg)
		case "Привязанный перевод":
			msg := tgbotapi.NewMessage(chatID, "")
			result := client.GetAllExpensesByGroupUtil(api, userID, msg, userChoiceState)
			if msg.Text != "" {
				bot.Send(msg)
				break
			}
			if len(result) == 0 {
				msg.Text = "Траты не найдены - создайте!!"
				bot.Send(msg)
				break
			}
			msg.ParseMode = "MarkdownV2"
			for i, expense := range result {
				msg.Text += fmt.Sprintf(
					"%d\\.\n"+
						"Название траты: %s\n"+
						"Общая сумма: %d\n"+
						"ID траты: `%s`\n",
					i+1,
					*expense.Title,
					int(*expense.TotalAmount),
					expense.Id.String(),
				)
			}
			msg.Text += "\n\nВведите ID траты:"
			userStates[userID] = "waiting_payment_expense_users"
			bot.Send(msg)

		case "Свободный перевод":
			msg := tgbotapi.NewMessage(chatID, "")
			respGroups, err := api.GetApiGroupsGroupId(context.Background(), userChoiceState[userID])
			if err != nil {
				msg.Text = "API недоступно"
				bot.Send(msg)
				break
			}
			defer respGroups.Body.Close()

			if respGroups.StatusCode != http.StatusOK && respGroups.StatusCode != http.StatusCreated {
				body, _ := io.ReadAll(respGroups.Body)
				fmt.Errorf("unexpected status code: %d, body: %s", respGroups.StatusCode, string(body))
				msg.Text = "Не удалось получить ответ от сервера"
				bot.Send(msg)
				break
			}

			var resultGroups swagger.GroupResponseDto
			if err := json.NewDecoder(respGroups.Body).Decode(&resultGroups); err != nil {
				fmt.Errorf("failed to decode response: %w", err)
				msg.Text = "Не удалось расшифровать ответ"
				bot.Send(msg)
				break
			}

			msg.Text = "Пользователи: \n\n"

			for i, member := range *resultGroups.Users {
				msg.Text += strconv.Itoa(i+1) + ". " + "Id: `" + member.Id.String() + "`\n Имя в системе: " + *member.DisplayName + "\n\n"
			}
			msg.Text += "\n\nВведите id человека, которому будете переводить деньги:"
			userStates[userID] = "waiting_payment_user_users"
			msg.ParseMode = "MarkdownV2"
			msg.Text = escapeMarkdown(msg.Text)
			bot.Send(msg)
		case "Изменить оплату":
			msg := tgbotapi.NewMessage(chatID, "")
			result := client.GetPaymentsByGroupIdUtil(api, userChoiceState[userID], msg)
			if len(result) == 0 {
				msg.Text += "У вас нет платежей - создайте!"
				bot.Send(msg)
				break
			}

			msg.Text = "Платежи: \n\n"

			for i, payment := range result {
				msg.Text += strconv.Itoa(i+1) + ":\n"
				msg.Text += "Id_перевода: `" + payment.Id.String() + "`\n"
				if payment.ExpenseId != nil {
					msg.Text += "От кого: " + payment.ExpenseId.String() + "\n"
				}
				if payment.FromUserName != nil {
					msg.Text += "От кого: " + *payment.FromUserName + "\n"
				}
				if payment.ToUserName != nil {
					msg.Text += "Кому: " + *payment.ToUserName + "\n"
				}
				msg.Text += "Сумма: " + strconv.Itoa(int(*payment.Amount)) + "\n"
			}
			msg.Text += "\n\nВведите id перевода который котите редактировать:"
			userStates[userID] = "waiting_payment"
			msg.ParseMode = "MarkdownV2"
			msg.Text = escapeMarkdown(msg.Text)
			bot.Send(msg)
		case "Удалить оплату":
			msg := tgbotapi.NewMessage(chatID, "")
			result := client.GetPaymentsByGroupIdUtil(api, userChoiceState[userID], msg)

			msg.Text = "Платежи: \n\n"

			for i, payment := range result {
				msg.Text += strconv.Itoa(i+1) + ":\n"
				msg.Text += "Id_траты: `" + payment.Id.String() + "`\n"
				if payment.ExpenseId != nil {
					msg.Text += "От кого: " + payment.ExpenseId.String() + "\n"
				}
				if payment.FromUserName != nil {
					msg.Text += "От кого: " + *payment.FromUserName + "\n"
				}
				if payment.ToUserName != nil {
					msg.Text += "Кому: " + *payment.ToUserName + "\n"
				}
				msg.Text += "Сумма: " + strconv.Itoa(int(*payment.Amount)) + "\n"
			}
			msg.Text += "\n\nВведите id перевода который котите редактировать:"
			userStates[userID] = "waiting_payment_to_delete"
			msg.ParseMode = "MarkdownV2"
			msg.Text = escapeMarkdown(msg.Text)
			bot.Send(msg)
		case "Получить список оплат":
			msg := tgbotapi.NewMessage(chatID, "")
			result := client.GetPaymentsByGroupIdUtil(api, userChoiceState[userID], msg)

			msg.Text = "Платежи: \n\n"

			for i, payment := range result {
				msg.Text += strconv.Itoa(i+1) + ":\n"
				msg.Text += "Id_траты: `" + payment.Id.String() + "`\n"
				if payment.ExpenseId != nil {
					msg.Text += "От кого: " + payment.ExpenseId.String() + "\n"
				}
				if payment.FromUserName != nil {
					msg.Text += "От кого: " + *payment.FromUserName + "\n"
				}
				if payment.ToUserName != nil {
					msg.Text += "Кому: " + *payment.ToUserName + "\n"
				}
				msg.Text += "Сумма: " + strconv.Itoa(int(*payment.Amount)) + "\n"
			}
			msg.ReplyMarkup = mainMenu
			msg.ParseMode = "MarkdownV2"
			msg.Text = escapeMarkdown(msg.Text)
			bot.Send(msg)

		default:
			//обработка состояний
			if state, ok := userStates[userID]; ok {
				switch state {
				case "waiting_group_selection":
					msg := tgbotapi.NewMessage(chatID, "")
					groupId, err := stringToUUID(update.Message.Text)
					if err != nil {
						msg.Text = "Не получилось обработать id"
						bot.Send(msg)
						break
					}

					resp, err := api.GetApiGroupsGroupId(context.Background(), groupId)
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

					var result swagger.GroupResponseDto
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						fmt.Errorf("failed to decode response: %w", err)
						msg.Text = "Не удалось расшифровать ответ"
						bot.Send(msg)
						break
					}

					msg.Text = fmt.Sprintf("Название группы: %s \nПользователи: \n", *result.Title)

					for i, member := range *result.Users {
						msg.Text += strconv.Itoa(i+1) + ". " + "Id: `" + member.Id.String() + "`\n Имя в системе: " + *member.DisplayName + "\n"
					}
					msg.ParseMode = "MarkdownV2"
					msg.Text = escapeMarkdown(msg.Text)
					bot.Send(msg)
				case "waiting_balance_group_selection":
					msg := tgbotapi.NewMessage(chatID, "")
					groupId, err := stringToUUID(update.Message.Text)
					if err != nil {
						msg.Text = "Не получилось обработать id"
						bot.Send(msg)
						break
					}

					resp, err := api.GetApiGroupsGroupIdBalance(context.Background(), groupId)
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

					var result swagger.BalanceResponseDto
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						fmt.Errorf("failed to decode response: %w", err)
						msg.Text = "Не удалось расшифровать ответ"
						bot.Send(msg)
						break
					}

					for i, balance := range *result.Balances {
						msg.Text += strconv.Itoa(i+1) + ": \n"
						msg.Text += "Id юзера: `" + *balance.DisplayName + "`\n"
						msg.Text += "Баланс: " + strconv.Itoa(int(*balance.Balance)) + "\n\n"
					}
					msg.ParseMode = "MarkdownV2"
					msg.Text = escapeMarkdown(msg.Text)
					bot.Send(msg)
				case "waiting_transfer_group_selection":
					msg := tgbotapi.NewMessage(chatID, "")
					groupId, err := stringToUUID(update.Message.Text)
					if err != nil {
						msg.Text = "Не получилось обработать id"
						bot.Send(msg)
						break
					}
					boolParam := true
					params := client.GetApiGroupsGroupIdTransfersParams{
						UseNpAlgorithm: &boolParam,
					}
					resp, err := api.GetApiGroupsGroupIdTransfers(context.Background(), groupId, &params)
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

					var result swagger.TransferSuggestionsResponseDto
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						fmt.Errorf("failed to decode response: %w", err)
						msg.Text = "Не удалось расшифровать ответ"
						bot.Send(msg)
						break
					}

					for i, transfer := range *result.Transfers {
						msg.Text += strconv.Itoa(i+1) + ": \n"
						msg.Text += "Имя отправителя: " + *transfer.FromUserName + "\n"
						msg.Text += "Имя получателя: " + *transfer.ToUserName + "\n"
						msg.Text += "Баланс: " + strconv.Itoa(int(*transfer.Amount)) + "\n\n"
					}
					msg.ParseMode = "MarkdownV2"
					msg.Text = escapeMarkdown(msg.Text)
					bot.Send(msg)
				case "waiting_expense_group_selection":
					msg := tgbotapi.NewMessage(chatID, "")
					groupId, err := stringToUUID(update.Message.Text)
					if err != nil {
						msg.Text = "Не получилось обработать id"
						bot.Send(msg)
						break
					}

					resp, err := api.GetApiGroupsGroupId(context.Background(), groupId)
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

					var result swagger.GroupResponseDto
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						fmt.Errorf("failed to decode response: %w", err)
						msg.Text = "Не удалось расшифровать ответ"
						bot.Send(msg)
						break
					}

					userChoiceState[userID] = groupId

					expenseMenu := tgbotapi.NewReplyKeyboard(
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton("Добавить трату"),
							tgbotapi.NewKeyboardButton("Изменить трату"),
						),
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton("Вернуться в меню"),
						),
					)
					msg = tgbotapi.NewMessage(chatID, "Выберите действие с тратами:")
					msg.ReplyMarkup = expenseMenu
					bot.Send(msg)
				case "waiting_expense_selection":
					msg := tgbotapi.NewMessage(chatID, "")
					expenseId, err := stringToUUID(update.Message.Text)
					if err != nil {
						msg.Text = "Не получилось обработать id"
						bot.Send(msg)
						break
					}
					result := client.GetExpenseByIdUtil(api, userID, expenseId, msg, userChoiceState)
					if msg.Text != "" {
						bot.Send(msg)
						break
					}

					userExpenceCreated[userID] = expenseId

					menuExpenseEdit := tgbotapi.NewReplyKeyboard(
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton("Изменить название траты"),
							tgbotapi.NewKeyboardButton("Изменить сумму траты"),
						),
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton("Добавить участников"),
							tgbotapi.NewKeyboardButton("Удалить трату"),
						),
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton("Вернуться в меню"),
						),
					)
					msg = tgbotapi.NewMessage(chatID, "Подробная информация о трате: \n")
					msg.Text += "Название траты: " + *result.Title + "\n"
					total_amount := int(*result.TotalAmount)
					msg.Text += "Сумма траты: " + strconv.Itoa(total_amount) + "\n"
					msg.Text += "Пользователи: \n"

					for i, user := range *result.Shares {
						msg.Text += strconv.Itoa(i+1) + ": \n"
						msg.Text += "Имя человека: `" + *user.UserName + "`\n"
						msg.Text += "Сколько должен: " + strconv.Itoa(int(*user.Amount)) + "\n"
					}
					msg.Text += "\n" + "Выберите действие с тратами:"
					msg.ReplyMarkup = menuExpenseEdit
					msg.ParseMode = "MarkdownV2"
					msg.Text = escapeMarkdown(msg.Text)
					bot.Send(msg)

				case "waiting_expense_title":
					userChoiceTitleState[userID] = update.Message.Text
					userStates[userID] = "waiting_expense_amount"
					bot.Send(tgbotapi.NewMessage(chatID, "Введите общую сумму траты"))

				case "waiting_expense_amount":

					msg := tgbotapi.NewMessage(chatID, "")
					creatorId := client.GetUserUUIDbyid(api, userID, msg)
					if msg.Text != "" {
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
					resp, err := api.PostApiExpensesGroupGroupId(context.Background(), userChoiceState[userID], reqBody)
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

					respGroups, err := api.GetApiGroupsGroupId(context.Background(), userChoiceState[userID])
					if err != nil {
						msg.Text = "API недоступно"
						bot.Send(msg)
						break
					}
					defer respGroups.Body.Close()

					if respGroups.StatusCode != http.StatusOK && respGroups.StatusCode != http.StatusCreated {
						body, _ := io.ReadAll(respGroups.Body)
						fmt.Errorf("unexpected status code: %d, body: %s", respGroups.StatusCode, string(body))
						msg.Text = "Не удалось получить ответ от сервера"
						bot.Send(msg)
						break
					}

					var resultGroups swagger.GroupResponseDto
					if err := json.NewDecoder(respGroups.Body).Decode(&resultGroups); err != nil {
						fmt.Errorf("failed to decode response: %w", err)
						msg.Text = "Не удалось расшифровать ответ"
						bot.Send(msg)
						break
					}

					msg.Text = fmt.Sprintf("Название группы: %s \n Id группы: `%s` \n Пользователи: \n", *resultGroups.Title, *resultGroups.Id)

					for i, member := range *resultGroups.Users {
						msg.Text += strconv.Itoa(i+1) + ". " + "Id: `" + member.Id.String() + "`\n Имя в системе: " + *member.DisplayName + "\n\n"
					}
					msg.Text += "Введите через пробел сумму которую должен человек и id человека:"
					msg.ParseMode = "MarkdownV2"
					msg.Text = escapeMarkdown(msg.Text)
					bot.Send(msg)

					userStates[userID] = "waiting_expense_users"
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
					idGroup := openapi_types.UUID(userChoiceState[userID])
					params := client.PostApiExpensesExpenseIdParticipantsParams{
						GroupId: &idGroup,
					}
					resp, err := api.PostApiExpensesExpenseIdParticipants(context.Background(), userExpenceCreated[userID], &params, reqBody)
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
					menu := tgbotapi.NewReplyKeyboard(
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton("Вернуться в меню"),
						),
					)
					msg.Text = "Пользователь добавлен в трату, если хотите закончить добавление - нажмите на кнопку Вернуться в меню"
					msg.ReplyMarkup = menu
					bot.Send(msg)
				case "waiting_new_expense_total_amount_users":
					msg := tgbotapi.NewMessage(chatID, "")
					reqBody, err := strconv.ParseFloat(update.Message.Text, 64)
					if err != nil {
						msg.Text = "Некорректный ввод суммы траты"
						bot.Send(msg)
						break
					}
					resp, err := api.PutApiExpensesExpenseIdTotalAmount(context.Background(), userExpenceCreated[userID], reqBody)
					if err != nil {
						msg.Text = "Не получилось изменить сумму траты"
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

					msg.Text = "Общая сумма траты успешно изменена"
					bot.Send(msg)
				case "waiting_username_selection":
					msg := tgbotapi.NewMessage(chatID, "")
					reqBody := client.UpdateUserRequestDto{
						DisplayName: &update.Message.Text,
					}
					resp, err := api.PutApiUsersUserId(context.Background(), userSysID[userID], reqBody)
					if err != nil {
						msg.Text = "Не получилось изменить сумму траты"
						bot.Send(msg)
						break
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusNoContent {
						body, _ := io.ReadAll(resp.Body)
						fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
						msg.Text = "Не удалось получить ответ от сервера"
						bot.Send(msg)
						break
					}

					msg.Text = "Имя успешно изменено"
					bot.Send(msg)

				case "waiting_group_edit_name_selection":
					msg := tgbotapi.NewMessage(chatID, "")
					reqBody := client.UpdateGroupRequestDto{
						Title: &update.Message.Text,
					}
					resp, err := api.PutApiGroupsGroupId(context.Background(), userChoiceState[userID], reqBody)
					if err != nil {
						msg.Text = "Не получилось изменить сумму траты"
						bot.Send(msg)
						break
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						body, _ := io.ReadAll(resp.Body)
						fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
						msg.Text = "Не удалось получить ответ от сервера"
						bot.Send(msg)
						break
					}
					msg.Text = "Успешно изменена группа!"
					msg.ReplyMarkup = mainMenu
					bot.Send(msg)
				case "waiting_group_edit_selection":
					msg := tgbotapi.NewMessage(chatID, "")
					groupId, err := stringToUUID(update.Message.Text)
					if err != nil {
						msg.Text = "Не получилось обработать id"
						bot.Send(msg)
						break
					}
					userChoiceState[userID] = groupId
					userStates[userID] = "waiting_group_edit_name_selection"
					msg.Text = "Выбирите название для группы"
					bot.Send(msg)
				case "waiting_group_delete_selection":
					msg := tgbotapi.NewMessage(chatID, "")
					groupId, err := stringToUUID(update.Message.Text)
					if err != nil {
						msg.Text = "Не получилось обработать id"
						bot.Send(msg)
						break
					}
					resp, err := api.DeleteApiGroupsGroupId(context.Background(), groupId)
					if err != nil {
						msg.Text = "Не получилось изменить сумму траты"
						bot.Send(msg)
						break
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						body, _ := io.ReadAll(resp.Body)
						fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
						msg.Text = "Не удалось получить ответ от сервера"
						bot.Send(msg)
						break
					}
					msg.Text = "Группа удалена!"
					msg.ReplyMarkup = mainMenu
					bot.Send(msg)
				case "waiting_group_id_join":
					msg := tgbotapi.NewMessage(chatID, "")
					groupId, err := stringToUUID(update.Message.Text)
					if err != nil {
						msg.Text = "Не получилось обработать id"
						bot.Send(msg)
						break
					}
					reqBody := client.AddGroupUserRequestDto{
						TelegramId: &userID,
					}
					resp, err := api.PostApiGroupsGroupIdUsers(context.Background(), groupId, reqBody)
					if err != nil {
						msg.Text = "Не получилось изменить сумму траты"
						bot.Send(msg)
						break
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						body, _ := io.ReadAll(resp.Body)
						fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
						msg.Text = "Не удалось получить ответ от сервера"
						bot.Send(msg)
						break
					}

					msg.Text = "Вы успешно добавлены в группу!"
					msg.ReplyMarkup = mainMenu
					bot.Send(msg)

				case "waiting_group_name_create_selection":
					msg := tgbotapi.NewMessage(chatID, "")
					reqBody := client.CreateGroupRequestDto{
						CreatedByTelegramId: &userID,
						TelegramChatId:      nil,
						Title:               &update.Message.Text,
					}
					respGroups, err := api.PostApiGroups(context.Background(), reqBody)
					if err != nil {
						msg.Text = "API недоступно"
						bot.Send(msg)
						break
					}
					defer respGroups.Body.Close()

					if respGroups.StatusCode != http.StatusOK && respGroups.StatusCode != http.StatusCreated {
						body, _ := io.ReadAll(respGroups.Body)
						fmt.Errorf("unexpected status code: %d, body: %s", respGroups.StatusCode, string(body))
						msg.Text = "Не удалось получить ответ от сервера"
						bot.Send(msg)
						break
					}

					msg.Text = "Группа успешно создана!"
					msg.ReplyMarkup = mainMenu
					bot.Send(msg)
				case "waiting_payment_to_delete":
					msg := tgbotapi.NewMessage(chatID, "")
					paymentId, err := stringToUUID(update.Message.Text)
					if err != nil {
						msg.Text = "Не получилось обработать id"
						bot.Send(msg)
						break
					}
					respGroups, err := api.DeleteApiGroupsGroupIdPaymentsPaymentId(context.Background(), userChoiceState[userID], paymentId)
					if err != nil {
						msg.Text = "API недоступно"
						bot.Send(msg)
						break
					}
					defer respGroups.Body.Close()

					if respGroups.StatusCode != http.StatusOK && respGroups.StatusCode != http.StatusCreated {
						body, _ := io.ReadAll(respGroups.Body)
						fmt.Errorf("unexpected status code: %d, body: %s", respGroups.StatusCode, string(body))
						msg.Text = "Не удалось получить ответ от сервера"
						bot.Send(msg)
						break
					}

					msg.Text = "Платеж успешно удален!"
					msg.ReplyMarkup = mainMenu
					bot.Send(msg)
				case "waiting_payment":
					msg := tgbotapi.NewMessage(chatID, "")
					paymentId, err := stringToUUID(update.Message.Text)
					if err != nil {
						msg.Text = "Не получилось обработать id"
						bot.Send(msg)
						break
					}

					userPaymentID[userID] = paymentId
					msg.Text = "Введите новую сумму платежа"
					bot.Send(msg)
					userStates[userID] = "waiting_payment_amount_edit"

				case "waiting_payment_amount_edit":
					msg := tgbotapi.NewMessage(chatID, "")
					amount, err := strconv.ParseFloat(update.Message.Text, 64)
					if err != nil {
						msg.Text = "Некорректный ввод, попробуйте заново"
						bot.Send(msg)
						break
					}
					reqBody := client.UpdatePaymentRequestDto{
						Amount: &amount,
					}
					respGroups, err := api.PutApiGroupsGroupIdPaymentsPaymentId(context.Background(), userChoiceState[userID], userPaymentID[userID], reqBody)
					if err != nil {
						msg.Text = "API недоступно"
						bot.Send(msg)
						break
					}
					defer respGroups.Body.Close()

					if respGroups.StatusCode != http.StatusOK && respGroups.StatusCode != http.StatusCreated {
						body, _ := io.ReadAll(respGroups.Body)
						fmt.Errorf("unexpected status code: %d, body: %s", respGroups.StatusCode, string(body))
						msg.Text = "Не удалось получить ответ от сервера"
						bot.Send(msg)
						break
					}

					msg.Text = "Платеж успешно изменен!"
					msg.ReplyMarkup = mainMenu
					bot.Send(msg)

				case "waiting_payment_user_users":
					msg := tgbotapi.NewMessage(chatID, "")
					userId, err := stringToUUID(update.Message.Text)
					if err != nil {
						msg.Text = "Не получилось обработать id"
						bot.Send(msg)
						break
					}

					userSysID[userID] = userId
					msg.Text = "Введите сумму платежа"
					bot.Send(msg)
					userStates[userID] = "waiting_payment_amount_users"

				case "waiting_payment_amount_users":
					msg := tgbotapi.NewMessage(chatID, "")
					amount, err := strconv.ParseFloat(update.Message.Text, 64)
					if err != nil {
						msg.Text = "Некорректный ввод, попробуйте заново"
						bot.Send(msg)
						break
					}

					userSysId := client.GetUserUUIDbyid(api, userID, msg)
					if msg.Text != "" {
						bot.Send(msg)
						break
					}
					toUserId := userSysID[userID]
					reqBody := client.CreateDirectPaymentRequestDto{
						Amount:     &amount,
						FromUserId: userSysId,
						ToUserId:   &toUserId,
					}
					respGroups, err := api.PostApiGroupsGroupIdPaymentsDirect(context.Background(), userChoiceState[userID], reqBody)
					if err != nil {
						msg.Text = "API недоступно"
						bot.Send(msg)
						break
					}
					defer respGroups.Body.Close()

					if respGroups.StatusCode != http.StatusOK && respGroups.StatusCode != http.StatusCreated {
						body, _ := io.ReadAll(respGroups.Body)
						fmt.Errorf("unexpected status code: %d, body: %s", respGroups.StatusCode, string(body))
						msg.Text = "Не удалось получить ответ от сервера"
						bot.Send(msg)
						break
					}

					var resultGroups swagger.PaymentResponseDto
					if err := json.NewDecoder(respGroups.Body).Decode(&resultGroups); err != nil {
						fmt.Errorf("failed to decode response: %w", err)
						msg.Text = "Не удалось расшифровать ответ"
						bot.Send(msg)
						break
					}

					msg.Text = "Платеж успешно создан!"
					msg.ReplyMarkup = mainMenu
					bot.Send(msg)
				case "waiting_payment_expense_users":
					msg := tgbotapi.NewMessage(chatID, "")
					expenseId, err := stringToUUID(update.Message.Text)
					if err != nil {
						msg.Text = "Не получилось обработать id"
						bot.Send(msg)
						break
					}
					_ = client.GetExpenseByIdUtil(api, userID, expenseId, msg, userChoiceState)
					if msg.Text != "" {
						bot.Send(msg)
						break
					}

					userExpenceCreated[userID] = expenseId
					msg.Text = "Введите сумму платежа"
					bot.Send(msg)
					userStates[userID] = "waiting_payment_expense_amount_users"

				case "waiting_3rd_person_expense_user":
					msg := tgbotapi.NewMessage(chatID, "")
					userId, err := stringToUUID(update.Message.Text)
					if err != nil {
						msg.Text = "Не получилось обработать id"
						bot.Send(msg)
						break
					}
					userSysID[userID] = userId
					userChoiceTitleState[userID] = update.Message.Text
					userStates[userID] = "waiting_expense_title_3rd"
					bot.Send(tgbotapi.NewMessage(chatID, "Введите название траты"))

				case "waiting_expense_title_3rd":
					userChoiceTitleState[userID] = update.Message.Text
					userStates[userID] = "waiting_expense_amount_3rd"
					bot.Send(tgbotapi.NewMessage(chatID, "Введите общую сумму траты"))

				case "waiting_expense_amount_3rd":

					msg := tgbotapi.NewMessage(chatID, "")
					if msg.Text != "" {
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
					userIdCreator := userSysID[userID]
					reqBody := client.CreateExpenseRequestDto{
						CreatedById: &userIdCreator,
						IsDraft:     &isDraft,
						Title:       &Title,
						TotalAmount: &TotalAmountFloat,
						Shares:      &SharesEmpty,
					}
					resp, err := api.PostApiExpensesGroupGroupId(context.Background(), userChoiceState[userID], reqBody)
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

					respGroups, err := api.GetApiGroupsGroupId(context.Background(), userChoiceState[userID])
					if err != nil {
						msg.Text = "API недоступно"
						bot.Send(msg)
						break
					}
					defer respGroups.Body.Close()

					if respGroups.StatusCode != http.StatusOK && respGroups.StatusCode != http.StatusCreated {
						body, _ := io.ReadAll(respGroups.Body)
						fmt.Errorf("unexpected status code: %d, body: %s", respGroups.StatusCode, string(body))
						msg.Text = "Не удалось получить ответ от сервера"
						bot.Send(msg)
						break
					}

					var resultGroups swagger.GroupResponseDto
					if err := json.NewDecoder(respGroups.Body).Decode(&resultGroups); err != nil {
						fmt.Errorf("failed to decode response: %w", err)
						msg.Text = "Не удалось расшифровать ответ"
						bot.Send(msg)
						break
					}

					msg.Text = fmt.Sprintf("Название группы: %s \n Id группы: `%s` \n Пользователи: \n", *resultGroups.Title, *resultGroups.Id)

					for i, member := range *resultGroups.Users {
						msg.Text += strconv.Itoa(i+1) + ". " + "Id: `" + member.Id.String() + "`\n Имя в системе: " + *member.DisplayName + "\n\n"
					}
					msg.Text += "Введите через пробел сумму которую должен человек и id человека:"
					msg.ParseMode = "MarkdownV2"
					msg.Text = escapeMarkdown(msg.Text)
					bot.Send(msg)

					userStates[userID] = "waiting_expense_users"

				case "waiting_payment_expense_amount_users":
					msg := tgbotapi.NewMessage(chatID, "")
					amount, err := strconv.ParseFloat(update.Message.Text, 64)
					if err != nil {
						msg.Text = "Некорректный ввод суммы траты, попробуйте заново"
						bot.Send(msg)
						break
					}

					userSysId := client.GetUserUUIDbyid(api, userID, msg)
					if msg.Text != "" {
						bot.Send(msg)
						break
					}

					expenseId := userExpenceCreated[userID]
					reqBody := client.CreatePaymentForExpenseRequestDto{
						Amount:     &amount,
						ExpenseId:  &expenseId,
						FromUserId: userSysId,
					}
					respGroups, err := api.PostApiGroupsGroupIdPaymentsExpense(context.Background(), userChoiceState[userID], reqBody)
					if err != nil {
						msg.Text = "API недоступно"
						bot.Send(msg)
						break
					}
					defer respGroups.Body.Close()

					if respGroups.StatusCode != http.StatusOK && respGroups.StatusCode != http.StatusCreated {
						body, _ := io.ReadAll(respGroups.Body)
						fmt.Errorf("unexpected status code: %d, body: %s", respGroups.StatusCode, string(body))
						msg.Text = "Не удалось получить ответ от сервера"
						bot.Send(msg)
						break
					}

					var resultGroups swagger.PaymentResponseDto
					if err := json.NewDecoder(respGroups.Body).Decode(&resultGroups); err != nil {
						fmt.Errorf("failed to decode response: %w", err)
						msg.Text = "Не удалось расшифровать ответ"
						bot.Send(msg)
						break
					}

					msg.Text = "Платеж успешно создан!"
					msg.ReplyMarkup = mainMenu
					bot.Send(msg)

				case "waiting_payments_group_selection":
					msg := tgbotapi.NewMessage(chatID, "")
					groupId, err := stringToUUID(update.Message.Text)
					if err != nil {
						msg.Text = "Не получилось обработать id"
						bot.Send(msg)
						break
					}

					resp, err := api.GetApiGroupsGroupId(context.Background(), groupId)
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

					var result swagger.GroupResponseDto
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						fmt.Errorf("failed to decode response: %w", err)
						msg.Text = "Не удалось расшифровать ответ"
						bot.Send(msg)
						break
					}

					userChoiceState[userID] = groupId

					paymentMenu := tgbotapi.NewReplyKeyboard(
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton("Создать оплату"),
							tgbotapi.NewKeyboardButton("Изменить оплату"),
						),
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton("Получить список оплат"),
							tgbotapi.NewKeyboardButton("Удалить оплату"),
						),
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton("Вернуться в меню"),
						),
					)
					msg = tgbotapi.NewMessage(chatID, "Выберите действие с переводами:")
					msg.ReplyMarkup = paymentMenu
					bot.Send(msg)

				case "waiting_new_expense_title_users":
					msg := tgbotapi.NewMessage(chatID, "")
					reqBody := update.Message.Text
					resp, err := api.PutApiExpensesExpenseIdTitle(context.Background(), userExpenceCreated[userID], reqBody)
					if err != nil {
						msg.Text = "Не получилось изменить название траты"
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

					msg.Text = "Название траты успешно изменено"
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

func escapeMarkdown(text string) string {
	escapeChars := []string{"_", "*", "[", "]", "(", ")", "~", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	for _, char := range escapeChars {
		text = strings.ReplaceAll(text, char, "\\"+char)
	}
	return text
}
