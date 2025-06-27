package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/msLoginoffTeam/tg_splitter_adapter/handles"
	tgutils "github.com/msLoginoffTeam/tg_splitter_adapter/handles/tg_utils"
	client "github.com/msLoginoffTeam/tg_splitter_adapter/swagger"
)

func main() {
	token := os.Getenv("BOT_TOKEN")

	log.Println("Bot Token:", token)

	if token == "" {
		log.Fatal("BOT_TOKEN is not set")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Начни работу с ботом"},
		{Command: "register", Description: "Зарегистрироваться в системе"},
		{Command: "creategroup", Description: "Создать новую группу"},
		{Command: "addtogroup", Description: "Добавить меня в группу: /команда id_группы"},
		{Command: "groupdetails", Description: "Информация о группе: /команда id_группы"},
		{Command: "renamegroup", Description: "Переименовать группу: /команда новое_название группы"},
		{Command: "help", Description: "Список команд"},
	}
	if _, err := bot.Request(tgbotapi.NewSetMyCommands(commands...)); err != nil {
		log.Printf("Ошибка установки команд: %v", err)
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := bot.GetUpdatesChan(updateConfig)

	// создание клиента для обращения к апишке
	baseUrl := os.Getenv("BACKEND_URL")
	log.Println("API URL:", baseUrl)
	if baseUrl == "" {
		log.Fatal("BACKEND_URL is not set")
	}

	apiClient, err := client.NewClientWithResponses(baseUrl)
	if err != nil {
		log.Panicf("Error creating API client: %v", err)
	}

	adapter := tgutils.NewCommandAdapter(tgutils.SimpleSplitter{})

	// Хранилище состояний пользователей
	userStates := make(map[int64]string)
	userChoiceState := make(map[int64]uuid.UUID)
	userChoiceTitleState := make(map[int64]string)
	userExpenceCreated := make(map[int64]uuid.UUID)

	for update := range updates {
		//if update.CallbackQuery != nil {
		//	// Обрабатываем callback-запросы
		//	handles.HandleCallbackQuery(bot, update.CallbackQuery, userStates, apiClient)
		//	continue
		//}

		if update.Message == nil {
			continue
		}

		if update.Message.Chat.IsPrivate() {
			//лс
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			handles.HandleDirectMessages(&update, bot, apiClient, adapter, userStates, userChoiceState, userChoiceTitleState, userExpenceCreated)
		} else if update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup() {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			//чат
			handles.HandleCommand(&update, bot, apiClient, adapter)
		}
	}
}
