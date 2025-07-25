package telegram

import (
	"KinopoiskTwoActors/internal/domain"
	"KinopoiskTwoActors/pkg/prometheus"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"strings"
	"time"
)

const (
	StepFirstActor        = "first_actor"
	StepFirstActorSelect  = "first_actor_select"
	StepSecondActor       = "second_actor"
	StepSecondActorSelect = "second_actor_select"
	StepCompleted         = "completed"
	correlationIDKey      = "correlation_id"
	chatIDKey             = "chat_id"
	commandKey            = "command"
	errorKey              = "error"
	successKey            = "success"
	queryKey              = "query"
	delay                 = time.Millisecond * 100
)

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	switch {
	case update.CallbackQuery != nil:
		b.handleCallback(ctx, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data,
			update.CallbackQuery.ID, update.CallbackQuery.Message.MessageID)

	case update.Message.IsCommand():
		b.handleCommand(ctx, update.Message.Chat.ID, update.Message.Command(),
			update.Message.CommandArguments())

	case update.Message == nil:
		return

	default:
		b.HandleSearchByTwoActors(ctx, update.Message.Chat.ID,
			strings.TrimSpace(update.Message.Text))
	}
}

func (b *Bot) handleCommand(ctx context.Context, chatID int64, command string, query string) {
	startTime := time.Now()
	defer func() {
		prometheus.CommandDuration.WithLabelValues(command).Observe(time.Since(startTime).Seconds())
	}()

	status := successKey
	defer func() {
		prometheus.CommandCounter.WithLabelValues(command, status).Inc()
	}()

	ctx = context.WithValue(ctx, correlationIDKey, b.GetCorrelationID(ctx, chatID))

	b.log.Info(
		"Команда получена", chatIDKey, chatID, commandKey, command, queryKey, query,
		correlationIDKey, ctx.Value(correlationIDKey))

	switch command {
	case "start":
		b.handleStart(ctx, chatID)
	case "help":
		b.handleHelp(ctx, chatID)
	default:
		status = errorKey
		b.handleUnknown(ctx, chatID)
	}
}

func (b *Bot) handleStart(ctx context.Context, chatID int64) {
	state := b.GetStateByID(ctx, chatID)
	*state = domain.SessionState{
		Step: StepFirstActor,
	}
	err := b.SetState(ctx, chatID, state)
	if err != nil {
		b.log.Error(
			"Ошибка задания шага",
			chatIDKey, chatID,
			correlationIDKey, ctx.Value(correlationIDKey),
			errorKey, err)
	}
	prometheus.ActiveUsers.Inc()
	b.SendMessage(ctx, chatID, "Введите имя первого актера")
}

func (b *Bot) handleHelp(ctx context.Context, chatID int64) {
	b.SendMessage(ctx, chatID, "Бот позволяет найти общие фильмы для двух актеров.\n"+
		"Для начала поиска нажмите /start")
}

func (b *Bot) handleUnknown(ctx context.Context, chatID int64) {
	b.SendMessage(ctx, chatID, "Неизвестная команда.\nВведите /start для нового поиска")
}

func (b *Bot) HandleSearchByTwoActors(ctx context.Context, chatID int64, query string) {
	state := b.GetStateByID(ctx, chatID)
	startTime := time.Now()
	defer func() {
		prometheus.CommandDuration.WithLabelValues(state.Step).Observe(time.Since(startTime).
			Seconds())
	}()
	ctx = context.WithValue(ctx, correlationIDKey, b.GetCorrelationID(ctx, chatID))

	status := successKey
	defer func() {
		prometheus.CommandCounter.WithLabelValues("search", status).Inc()
	}()

	switch state.Step {
	case StepFirstActor, StepSecondActor:
		err := b.handleActor(ctx, chatID, query)
		if err != nil {
			status = errorKey
			b.log.Error(
				"Ошибка обработки поиска актера",
				chatIDKey, chatID,
				queryKey, query,
				correlationIDKey, ctx.Value(correlationIDKey),
				errorKey, err)
			b.ResetUserState(ctx, chatID)
			b.SendMessage(ctx, chatID, "Произошла ошибка поиска. Введите /start для нового поиска")
		}
		b.log.Info(
			"Актеры успешно отправлены на выбор",
			chatIDKey, chatID,
			queryKey, query,
			correlationIDKey, ctx.Value(correlationIDKey),
		)
	default:
		b.SendMessage(ctx, chatID, "Введите /start для нового поиска")
		b.log.Debug(
			"Ошибка шага",
			chatIDKey, chatID,
			"state.Step", state.Step,
			queryKey, query,
			correlationIDKey, ctx.Value(correlationIDKey),
		)
	}
}

func (b *Bot) handleActor(ctx context.Context, chatID int64, query string) error {
	const op = "BotHandler.handleActor"

	state := b.GetStateByID(ctx, chatID)

	actors, err := b.SearchActor(ctx, query)
	if err != nil {
		return fmt.Errorf("%s: Ошибка поиска актера %s: %w", op, query, err)
	}

	if len(actors) == 0 {
		b.log.Info("Актеры не найдены", chatIDKey, chatID, queryKey, query, correlationIDKey,
			ctx.Value(correlationIDKey))
		return fmt.Errorf("%s: Актеры по запросу \"%s\"не найдены", op, query)
	}

	state.TempActors = b.createPhotoData(actors)

	if state.Step == StepFirstActor {
		state.Step = StepFirstActorSelect
	} else if state.Step == StepSecondActor {
		state.Step = StepSecondActorSelect
	} else {
		state.Step = StepCompleted
	}

	b.log.Debug("Подготовлены к отправке на выбор:",
		"state.TempActors", state.TempActors,
		chatIDKey, chatID,
		correlationIDKey, ctx.Value(correlationIDKey),
	)

	err = b.sendActors(ctx, chatID, state.TempActors)
	if err != nil {
		return fmt.Errorf("%s: Ошибка отправки актеров на выбор %s: %w", op, query, err)
	}

	return nil
}

func (b *Bot) sendActors(ctx context.Context, chatID int64, actors []domain.PhotoData) error {
	const op = "BotHandler.sendActors"

	b.SendMessage(ctx, chatID, "Найдены")

	for _, photo := range actors {
		if _, err := b.SendActorWithPhoto(ctx, chatID, photo); err != nil {
			return fmt.Errorf("%s: ошибка отправки фото в чат %d: %v", op, chatID, err)
		}
		time.Sleep(delay)
	}

	return nil
}

func (b *Bot) SendActorWithPhoto(ctx context.Context, chatID int64,
	photo domain.PhotoData) (int, error) {
	data := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(photo.PhotoURL))
	data.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Ссылка", photo.ActorURL),
			tgbotapi.NewInlineKeyboardButtonData("Выбрать",
				strconv.Itoa(photo.ID)),
		),
	)
	data.Caption = photo.Caption
	sentMsg, err := b.Send(data)
	if err != nil {
		return 0, err
	}
	state := b.GetStateByID(ctx, chatID)
	state.SentMediaMessages = append(state.SentMediaMessages, sentMsg.MessageID)
	return sentMsg.MessageID, nil
}

func (b *Bot) handleActorSelection(ctx context.Context, chatID int64, actorID int) {
	state := b.GetStateByID(ctx, chatID)
	if err := b.ClearPreviousMedia(ctx, chatID); err != nil {
		b.log.Error("Ошибка очистки медиа", err, chatIDKey, chatID, correlationIDKey,
			ctx.Value(correlationIDKey))
	}

	switch state.Step {
	case StepFirstActorSelect:
		state.FirstActorID = actorID
		state.Step = StepSecondActor
		b.SendMessage(ctx, chatID, "Введите имя второго актера:")

	case StepSecondActorSelect:
		state.SecondActorID = actorID
		state.Step = StepCompleted
		err := b.handleCommonMovies(ctx, chatID, state)
		if err != nil {
			b.ResetUserState(ctx, chatID)
			b.log.Error("Ошибка обработки вывода фильмов", err, chatIDKey, chatID,
				correlationIDKey, ctx.Value(correlationIDKey))
		}
	default:
		b.SendMessage(ctx, chatID, "Неверный выбор. Введите /start")
	}
}
func (b *Bot) handleCallback(ctx context.Context, chatID int64, data string, callbackID string,
	callbackMessageID int) {
	const op = "BotHandler.handleCallback"
	actorID, err := strconv.Atoi(data)
	if err != nil {
		b.log.Error(
			"Ошибка конвертации ID актера",
			chatIDKey, chatID,
			correlationIDKey, ctx.Value(correlationIDKey),
			errorKey, err)
		b.SendMessage(ctx, chatID, "Произошла ошибка поиска. Введите /start для нового поиска")
		b.ResetUserState(ctx, chatID)
		return
	}
	ctx = context.WithValue(ctx, correlationIDKey, b.GetCorrelationID(ctx, chatID))
	b.log.Info("Выбран актер", "actorID", actorID, chatIDKey, chatID, correlationIDKey,
		ctx.Value(correlationIDKey))
	b.handleActorSelection(ctx, chatID, actorID)
	var answerText string
	//if err != nil {
	//	answerText = "Ошибка выбора актера"
	//	log.Printf("Error selecting actor: %v", err)
	//
	//}
	err = b.AnswerCallbackQuery(callbackID, answerText)
	//if err != nil {
	//
	//}

	editMsg := tgbotapi.NewEditMessageReplyMarkup(
		chatID,
		callbackMessageID,
		tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("Ссылка",
					fmt.Sprintf("https://www.kinopoisk.ru/name/%d/", actorID)),
			),
		),
	)
	b.Send(editMsg)
}

func (b *Bot) handleCommonMovies(ctx context.Context, chatID int64, state *domain.SessionState) error {

	commonMovies, err := b.GetCommonMovies(ctx, state.FirstActorID, state.SecondActorID)

	if err != nil {
		return err
	}

	if len(commonMovies) == 0 {
		b.SendMessage(ctx, chatID, "У актеров нет общих фильмов")
	} else if len(commonMovies) > 10 {
		b.SendMessage(ctx, chatID, "Общих фильмов больше 10")
	} else {
		msg := fmt.Sprintf("Общие фильмы:")
		b.SendMessage(ctx, chatID, msg)

		for _, movie := range commonMovies {
			b.SendMovie(chatID, movie)
		}

	}
	b.ResetUserState(ctx, chatID)
	prometheus.ActiveUsers.Dec()
	return nil
}

func (b *Bot) createPhotoData(actors []domain.Actor) []domain.PhotoData {
	if len(actors) == 0 {
		return nil
	}

	response := make([]domain.PhotoData, 0, len(actors))
	for _, actor := range actors {
		birthday := time.Time{}
		if actor.Birthday != "" {
			var err error
			birthday, err = time.Parse(time.RFC3339, actor.Birthday)
			if err != nil {
				b.log.Debug("Ошибка парсинга даты", err, "actor.Birthday", actor.Birthday)
			}
		}
		photo := domain.PhotoData{
			ID:       actor.ID,
			PhotoURL: actor.PhotoURL,
			ActorURL: actor.ActorURL,
			Caption:  fmt.Sprintf("%s (%s), %d", actor.Name, actor.EngName, birthday.Year()),
		}
		response = append(response, photo)
	}
	return response
}

func (b *Bot) SendMovie(chatID int64, movie domain.Movie) error {
	const op = "BotHandler.sendMovie"

	data := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(movie.PosterURL))
	data.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Ссылка", movie.MovieURL),
		),
	)
	caption := fmt.Sprintf("%s (%s) %d, Рейтинг: %.1f", movie.Name, movie.EngName, movie.Year, movie.Rating)

	data.Caption = caption
	_, err := b.Send(data)
	return err
}
