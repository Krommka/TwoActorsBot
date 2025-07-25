package telegram

import (
	"KinopoiskTwoActors/internal/domain"
	"context"
)

type StateProvider interface {
	SetState(ctx context.Context, chatID int64, state *domain.SessionState) error
	GetStateByID(ctx context.Context, chatID int64) *domain.SessionState
	ResetUserState(ctx context.Context, chatID int64)
	GetCurrentStatesID(ctx context.Context) []int64
	GetCorrelationID(ctx context.Context, chatID int64) string
}

type ActorProvider interface {
	SearchActor(ctx context.Context, query string) ([]domain.Actor, error)
}

type FilmProvider interface {
	GetCommonMovies(ctx context.Context, firstActorID int, secondActorID int) ([]domain.Movie,
		error)
}
