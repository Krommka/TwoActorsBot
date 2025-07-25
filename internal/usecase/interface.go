package usecase

import (
	"KinopoiskTwoActors/internal/domain"
	"context"
)

type ActorFilmRepository interface {
	SearchActors(ctx context.Context, query string) ([]domain.Actor, error)
	GetMoviesIDByActorID(ctx context.Context, actorID int) ([]int, error)
	GetMovieByID(ctx context.Context, movieID int) (domain.Movie, error)
}
