package usecase

import (
	"KinopoiskTwoActors/internal/domain"
	"context"
	"fmt"
)

type Film struct {
	repo ActorFilmRepository
}

func NewFilm(repo ActorFilmRepository) *Film {
	return &Film{repo: repo}
}

func (uc *Film) GetCommonMovies(ctx context.Context, firstActorID int,
	secondActorID int) ([]domain.Movie, error) {

	if firstActorID == secondActorID {
		return nil, fmt.Errorf("актер задублирован")
	}

	commonMoviesID, err := uc.getCommonMoviesID(ctx, firstActorID, secondActorID)
	if err != nil {
		return nil, err
	}

	commonMovies := make([]domain.Movie, 0)

	for _, id := range commonMoviesID {
		movie, err := uc.repo.GetMovieByID(ctx, id)
		if err != nil {
			return nil, err
		}
		commonMovies = append(commonMovies, movie)
	}

	return commonMovies, nil
}

func (uc *Film) getCommonMoviesID(ctx context.Context, firstActorID int,
	secondActorID int) ([]int, error) {
	movies1, err := uc.repo.GetMoviesIDByActorID(ctx, firstActorID)
	if err != nil {
		return nil, err
	}
	movies2, err := uc.repo.GetMoviesIDByActorID(ctx, secondActorID)
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
