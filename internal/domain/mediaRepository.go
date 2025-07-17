package domain

type MediaRepository interface {
	SearchActors(query string) ([]Actor, error)
	GetMoviesIDByActorID(actorID int) ([]int, error)
	GetMovieByID(movieID int) (Movie, error)
}
