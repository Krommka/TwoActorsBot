package usecase

import "KinopoiskTwoActors/internal/domain"

type ActorFilmRepository struct {
	repo domain.MediaRepository
}

func NewActorFilmRepository(repo domain.MediaRepository) *ActorFilmRepository {
	return &ActorFilmRepository{repo}
}

func (uc *ActorFilmRepository) SearchActors(query string) ([]domain.Actor, error) {
	return uc.repo.SearchActors(query)
}

func (uc *ActorFilmRepository) GetMoviesIDByActorID(actorID int) ([]int, error) {
	return uc.repo.GetMoviesIDByActorID(actorID)
}

func (uc *ActorFilmRepository) GetMovieByID(movieID int) (domain.Movie, error) {
	return uc.repo.GetMovieByID(movieID)
}
