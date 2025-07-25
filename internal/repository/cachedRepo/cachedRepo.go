package cachedRepo

import (
	"KinopoiskTwoActors/internal/domain"
	"KinopoiskTwoActors/pkg/prometheus"
	"context"
	"errors"
	"fmt"
	"log/slog"
)

type ActorFilmRepository interface {
	SearchActors(ctx context.Context, query string) ([]domain.Actor, error)
	GetMoviesIDByActorID(ctx context.Context, actorID int) ([]int, error)
	GetMovieByID(ctx context.Context, movieID int) (domain.Movie, error)
}

type CacheRepository interface {
	GetMovieByID(ctx context.Context, movieID int) (domain.Movie, error)
	SetMovie(ctx context.Context, movie domain.Movie) error
}

type CachedRepo struct {
	repo  ActorFilmRepository
	cache CacheRepository
	log   *slog.Logger
}

func NewCachedRepo(repo ActorFilmRepository, cache CacheRepository, log *slog.Logger) *CachedRepo {

	return &CachedRepo{
		repo:  repo,
		cache: cache,
		log:   log,
	}
}

func (r *CachedRepo) SearchActors(ctx context.Context, query string) ([]domain.Actor, error) {
	return r.repo.SearchActors(ctx, query)
}
func (r *CachedRepo) GetMoviesIDByActorID(ctx context.Context, actorID int) ([]int, error) {
	return r.repo.GetMoviesIDByActorID(ctx, actorID)
}
func (r *CachedRepo) GetMovieByID(ctx context.Context, movieID int) (domain.Movie, error) {
	const op = "cachedRepo.GetMovieByID"
	movie, err := r.cache.GetMovieByID(ctx, movieID)
	if err == nil {
		prometheus.CacheOperations.WithLabelValues("hit").Inc()
		return movie, nil
	}
	if !errors.Is(err, domain.ErrRecordNotFound) {
		prometheus.CacheOperations.WithLabelValues("error").Inc()
		r.log.WarnContext(ctx, "cache lookup failed",
			"movieID", movieID,
			"error", err,
		)
	}
	prometheus.CacheOperations.WithLabelValues("miss").Inc()
	movie, err = r.repo.GetMovieByID(ctx, movieID)
	if err != nil {
		return domain.Movie{}, fmt.Errorf("%s: %w", op, err)
	}

	go func() {
		if err = r.cache.SetMovie(ctx, movie); err != nil {
			r.log.ErrorContext(ctx, "failed to cache movie",
				"movieID", movieID,
				"error", err,
			)
		}
	}()
	return movie, nil
}
