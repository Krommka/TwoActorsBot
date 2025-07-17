package telegram

import (
	"KinopoiskTwoActors/configs"
	"KinopoiskTwoActors/internal/repository/userState"
	"KinopoiskTwoActors/internal/usecase"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"log/slog"
	"net/http"
	"strings"
)

type Bot struct {
	*tgbotapi.BotAPI
	repo       *usecase.ActorFilmRepository
	userStates *userState.UserStates
	//mu         sync.RWMutex
	logger *slog.Logger
}

func NewBot(config *configs.Config, userStates *userState.UserStates, repo *usecase.ActorFilmRepository, log *slog.Logger) (*Bot, error) {

	api, err := tgbotapi.NewBotAPI(config.TG.Token)
	api.Client = &http.Client{
		Timeout: config.TG.ConnectionTimeout,
	}
	if err != nil {
		return nil, err
	}

	return &Bot{api, repo, userStates, log}, nil
}

func (b *Bot) wrapError(chatID int64, op, msg string, err error) error {
	sendErr := b.SendMessage(chatID, "Ошибка: "+msg)
	if sendErr != nil {
		log.Printf("%s: ошибка отправки сообщения: %v", op, sendErr)
	}
	if err != nil {
		return fmt.Errorf("%s: %s: %w", op, msg, err)
	}
	return fmt.Errorf("%s: %s", op, msg)
}

func (b *Bot) SendMessage(chatID int64, text string) error {
	if len(text) > 4000 {
		text = text[:4000] + "..."
	}
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.Send(msg)
	if err != nil {
		log.Printf("Ошибка отправки сообщения в чат %d: %v", chatID, err)
	}
	return err
}

func (b *Bot) Run() {
	u := tgbotapi.NewUpdate(0)
	updates := b.GetUpdatesChan(u)

	for update := range updates {
		switch {
		case update.CallbackQuery != nil:
			chatID := update.CallbackQuery.Message.Chat.ID
			data := update.CallbackQuery.Data
			b.HandleCallback(chatID, data, update.CallbackQuery.ID, update.CallbackQuery.Message.MessageID)

		case update.Message == nil:
			continue

		case update.Message.IsCommand():
			command := update.Message.Command()
			args := update.Message.CommandArguments()
			b.HandleCommand(update.Message.Chat.ID, command, args)

		default:
			text := strings.TrimSpace(update.Message.Text)
			b.HandleSearchByTwoActors(update.Message.Chat.ID, text)
		}
	}
}

func (b *Bot) AnswerCallbackQuery(callbackID string, text string) error {
	cfg := tgbotapi.NewCallback(callbackID, text)
	_, err := b.Request(cfg)
	return err
}
