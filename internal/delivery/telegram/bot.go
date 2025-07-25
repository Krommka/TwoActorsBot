package telegram

import (
	"KinopoiskTwoActors/configs"
	"KinopoiskTwoActors/pkg/prometheus"
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
	"net/http"
)

type Bot struct {
	*tgbotapi.BotAPI
	StateProvider
	ActorProvider
	FilmProvider
	log *slog.Logger
}

func NewBot(config *configs.Config, userStates StateProvider,
	actor ActorProvider, film FilmProvider, log *slog.Logger) (*Bot, error) {

	api, err := tgbotapi.NewBotAPI(config.TG.Token)
	if err != nil {
		return nil, err
	}
	api.Client = &http.Client{
		Timeout: config.TG.ConnectionTimeout,
	}

	return &Bot{api, userStates, actor, film, log}, nil
}

func (b *Bot) Run(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	updates := b.GetUpdatesChan(u)

	for update := range updates {
		select {
		case <-ctx.Done():
			return
		default:
			b.handleUpdate(ctx, update)
		}
	}
}

func (b *Bot) Stop(ctx context.Context) {
	ids := b.GetCurrentStatesID(ctx)
	for _, id := range ids {
		b.SendMessage(ctx, id, "Соединение разорвано")
	}
}

func (b *Bot) SendMessage(ctx context.Context, chatID int64, text string) {
	if len(text) > 4000 {
		text = text[:4000] + "..."
	}
	done := make(chan error, 1)
	msg := tgbotapi.NewMessage(chatID, text)

	go func() {
		_, err := b.Send(msg)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			prometheus.MessagesSent.WithLabelValues("error").Inc()
			b.log.Error("Ошибка отправки сообщения в чат",
				err,
				"text", text,
				chatIDKey, chatID,
				correlationIDKey, ctx.Value(correlationIDKey))
		} else {
			prometheus.MessagesSent.WithLabelValues("ok").Inc()
		}
	case <-ctx.Done():
		b.log.Error("Ошибка отправки сообщения в чат: context timeout",
			chatIDKey, chatID,
			correlationIDKey, ctx.Value(correlationIDKey))
	}

}

func (b *Bot) AnswerCallbackQuery(callbackID string, text string) error {
	cfg := tgbotapi.NewCallback(callbackID, text)
	_, err := b.Request(cfg)
	return err
}

func (b *Bot) DeleteMessage(chatID int64, messageID int) error {
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err := b.Request(deleteMsg)
	return err
}

func (b *Bot) ClearPreviousMedia(ctx context.Context, chatID int64) error {
	state := b.GetStateByID(ctx, chatID)
	b.log.Debug("Очистка предыдущих медиа", chatIDKey, chatID,
		correlationIDKey, ctx.Value(correlationIDKey))

	for _, msgID := range state.SentMediaMessages {
		if err := b.DeleteMessage(chatID, msgID); err != nil {
			b.log.Debug("Ошибка удаления сообщения", "msgID", msgID, chatIDKey, chatID,
				correlationIDKey, ctx.Value(correlationIDKey))
		}
	}
	state.SentMediaMessages = nil

	return nil
}
