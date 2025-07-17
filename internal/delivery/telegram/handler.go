package telegram

import (
	"KinopoiskTwoActors/internal/domain"
	"KinopoiskTwoActors/internal/repository/userState"
	"KinopoiskTwoActors/pkg/prometheus"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (b *Bot) getCorrelationID(chatID int64) string {
	state := b.userStates.GetUserState(chatID)
	if state.CorrelationID == "" {
		state.CorrelationID = generateCorrelationID()
	}
	return state.CorrelationID
}

func generateCorrelationID() string {
	return uuid.New().String()
}

func (b *Bot) HandleCommand(chatID int64, command string, query string) error {
	startTime := time.Now()
	defer func() {
		prometheus.CommandDuration.WithLabelValues(command).Observe(time.Since(startTime).Seconds())
	}()

	status := "success"
	defer func() {
		prometheus.CommandCounter.WithLabelValues(command, status).Inc()
	}()

	const op = "BotHandler.HandleCommand"
	ctx := context.WithValue(context.Background(), "correlation_id", generateCorrelationID())

	b.logger.Info(
		"Command received",
		"chat_id", chatID,
		"command", command,
		"correlation_id", ctx.Value("correlation_id").(string))

	switch command {
	case "start":
		return b.handleStart(chatID)
	case "help":
		return b.handleHelp(chatID)
	default:
		if err := b.SendMessage(chatID, "Неизвестная команда.\nВведите /start для нового поиска"); err != nil {
			log.Printf("%s: ошибка отправки сообщения в чат %d: %v", op, chatID, err)
			status = "error"
		}
		return nil
	}
}

func (b *Bot) handleStart(chatID int64) error {
	state := b.userStates.GetUserState(chatID)
	prometheus.ActiveUsers.Inc()
	*state = userState.State{
		Step: "first_actor",
	}
	return b.SendMessage(chatID, "Введите имя первого актера")
}

func (b *Bot) handleHelp(chatID int64) error {
	return b.SendMessage(chatID, "Бот позволяет найти общие фильмы для двух актеров.\n"+
		"Для начала поиска нажмите /start")
}

func (b *Bot) HandleSearchByTwoActors(chatID int64, query string) error {

	const op = "BotHandler.HandleCommand"
	state := b.userStates.GetUserState(chatID)

	switch state.Step {
	case "first_actor", "second_actor":
		return b.searchActor(chatID, query)
	default:
		return b.SendMessage(chatID, "Введите /start для нового поиска")
	}
}

func (b *Bot) HandleCallback(chatID int64, data string, callbackID string, callbackMessageID int) error {
	if _, err := strconv.Atoi(data); err == nil {
		actorID, _ := strconv.Atoi(data)
		err := b.handleActorSelection(chatID, actorID)
		if err != nil {
			log.Println(err)
		}
		answerText := ""
		if err != nil {
			answerText = "Ошибка выбора актера"
			log.Printf("Error selecting actor: %v", err)

		}
		b.AnswerCallbackQuery(callbackID, answerText)

		editMsg := tgbotapi.NewEditMessageReplyMarkup(
			chatID,
			callbackMessageID,
			tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonURL("Кинопоиск",
						fmt.Sprintf("https://www.kinopoisk.ru/name/%d/", actorID)),
				),
			),
		)
		b.Send(editMsg)
	}
	return nil
}

func (b *Bot) searchActor(chatID int64, query string) error {
	const op = "BotHandler.searchActor"

	state := b.userStates.GetUserState(chatID)

	if len(query) == 0 {
		return b.wrapError(chatID, op, "не указано имя", nil)
	}

	actors, err := b.repo.SearchActors(query)
	if err != nil {
		return b.wrapError(chatID, op, "ошибка поиска в usecase", err)
	}

	if len(actors) == 0 {
		return b.wrapError(chatID, op, "актеры по указанному запросу не найдены. Начните заново", nil)
	}

	normalizedQuery := normalizeName(query)
	filteredActors := make([]domain.Actor, 0)
	for _, actor := range actors {
		if normalizeName(actor.Name) == normalizedQuery ||
			normalizeName(actor.EngName) == normalizedQuery {
			filteredActors = append(filteredActors, actor)
			break
		}
	}

	if len(filteredActors) > 0 {
		state.TempActors = preparePhoto(filteredActors)
	} else {
		state.TempActors = preparePhoto(actors)
	}

	if state.Step == "first_actor" {
		state.Step = "first_actor_select"
	} else if state.Step == "second_actor" {
		state.Step = "second_actor_select"
	} else {
		state.Step = "complete"
	}

	return b.sendActorSelection(chatID, state.TempActors)

}

func normalizeName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	// Удаляем все не-буквенные символы для более гибкого сравнения
	reg := regexp.MustCompile(`[^a-zа-яё]`)
	return reg.ReplaceAllString(name, "")
}

func (b *Bot) sendActorSelection(chatID int64, actors []domain.PhotoData) error {
	const op = "BotHandler.sendActorSelection"

	if err := b.SendMessage(chatID, "Найдены"); err != nil {
		log.Printf("%s: ошибка отправки сообщения в чат %d: %v", op, chatID, err)
	}
	log.Printf("Исходный массив:%v", actors)

	for _, photo := range actors {
		time.Sleep(100 * time.Millisecond)
		if _, err := b.SendActorWithPhoto(chatID, photo); err != nil {
			log.Printf("%s: ошибка отправки фото в чат %d: %v", op, chatID, err)
		}
	}
	log.Printf("%s: Отправлен ответ в чат %d", op, chatID)
	return nil
}

func (b *Bot) handleActorSelection(chatID int64, actorID int) error {
	state := b.userStates.GetUserState(chatID)
	if err := b.ClearPreviousMedia(chatID); err != nil {
		log.Printf("Ошибка очистки медиа: %v", err)
	}

	switch state.Step {
	case "first_actor_select":
		state.FirstActorID = actorID
		state.Step = "second_actor"
		return b.SendMessage(chatID, "Введите имя второго актера:")

	case "second_actor_select":
		state.SecondActorID = actorID
		state.Step = "completed"
		return b.handleCommonMovies(chatID, state)

	default:
		return b.SendMessage(chatID, "Неверный выбор. Введите /start")
	}
}

func (b *Bot) handleCommonMovies(chatID int64, state *userState.State) error {

	if state.FirstActorID == state.SecondActorID {
		b.SendMessage(chatID, "Актер задублирован")
		return fmt.Errorf("актер задублирован")
	}

	commonMovies, err := b.getCommonMoviesID(state)
	if err != nil {
		return err
	}

	if len(commonMovies) == 0 {
		b.SendMessage(chatID, "У актеров нет общих фильмов")
	} else if len(commonMovies) > 10 {
		b.SendMessage(chatID, "Общих фильмов больше 10")
	} else {
		msg := fmt.Sprintf("Общие фильмы:")
		b.SendMessage(chatID, msg)

		for _, movieID := range commonMovies {
			b.SendMovie(chatID, movieID)
		}

	}
	b.userStates.ResetUserState(chatID)
	prometheus.ActiveUsers.Dec()
	return nil
}

func (b *Bot) getCommonMoviesID(state *userState.State) ([]int, error) {
	movies1, err := b.repo.GetMoviesIDByActorID(state.FirstActorID)
	if err != nil {
		return nil, err
	}
	movies2, err := b.repo.GetMoviesIDByActorID(state.SecondActorID)
	if err != nil {
		return nil, err
	}

	commonMovies := findCommonMoviesID(movies1, movies2)
	return commonMovies, nil
}

func findCommonMoviesID(movies1, movies2 []int) []int {
	if len(movies1) == 0 || len(movies2) == 0 {
		return nil
	}

	movieMap := make(map[int]bool)
	for _, movie := range movies1 {
		movieMap[movie] = true
	}

	common := make([]int, 0)
	for _, movie := range movies2 {
		if movieMap[movie] {
			common = append(common, movie)
			delete(movieMap, movie)
		}
	}
	return common
}

func preparePhoto(actors []domain.Actor) []domain.PhotoData {
	const op = "BotHandler.preparePhoto"

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
				log.Printf("%s: ошибка парсинга даты", op)
			}
		}

		if strings.HasPrefix(actor.Photo, "https:https://") {
			actor.Photo = strings.TrimPrefix(actor.Photo, "https:")
		}

		photo := domain.PhotoData{
			ID:      actor.ID,
			URL:     actor.Photo,
			Caption: fmt.Sprintf("%s (%s), %d", actor.Name, actor.EngName, birthday.Year()),
		}

		response = append(response, photo)
	}

	return response
}

func (b *Bot) SendMovie(chatID int64, movieID int) error {
	const op = "BotHandler.sendMovie"
	movie, err := b.repo.GetMovieByID(movieID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	data := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(movie.Poster))
	link := fmt.Sprintf("https://www.kinopoisk.ru/film/%d/", movie.ID)
	data.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Кинопоиск", link),
		),
	)
	caption := fmt.Sprintf("%s (%s) %d, Рейтинг: %.1f", movie.Name, movie.EngName, movie.Year, movie.Rating)

	data.Caption = caption
	_, err = b.Send(data)
	return err
}

func (b *Bot) SendActorWithPhoto(chatID int64, photo domain.PhotoData) (int, error) {
	log.Printf("Пытаюсь отправить фото по URL: %s", photo.URL)
	data := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(photo.URL))
	link := fmt.Sprintf("https://www.kinopoisk.ru/name/%d/", photo.ID)
	data.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Кинопоиск", link),
			tgbotapi.NewInlineKeyboardButtonData("Выбрать", strconv.Itoa(photo.ID)),
		),
	)
	data.Caption = photo.Caption
	sentMsg, err := b.Send(data)
	if err != nil {
		return 0, err
	}
	state := b.userStates.GetUserState(chatID)
	state.SentMediaMessages = append(state.SentMediaMessages, sentMsg.MessageID)
	return sentMsg.MessageID, nil
}

func (b *Bot) ClearPreviousMedia(chatID int64) error {
	state := b.userStates.GetUserState(chatID)

	for _, msgID := range state.SentMediaMessages {
		if err := b.DeletePhotoMessage(chatID, msgID); err != nil {
			log.Printf("Ошибка удаления сообщения %d: %v", msgID, err)
		}
	}
	state.SentMediaMessages = nil

	return nil
}

func (b *Bot) DeletePhotoMessage(chatID int64, messageID int) error {
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err := b.Request(deleteMsg)
	return err
}
