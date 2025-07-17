package kinopoisk

import (
	"KinopoiskTwoActors/configs"
	"KinopoiskTwoActors/internal/domain"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type Repo struct {
	Path   string
	APIKey string
	Client *http.Client
	log    *slog.Logger
}

func NewRepo(config *configs.Config) *Repo {

	return &Repo{
		APIKey: config.KP.Token,
		Path:   config.KP.Path,
		Client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (repo *Repo) GetMovieByID(movieID int) (domain.Movie, error) {
	req := fmt.Sprintf("movie/%d", movieID)

	resp, err := repo.doRequest(req)
	if err != nil {
		return domain.Movie{}, err
	}

	var movieInfo struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Rating struct {
			Kp   float32 `json:"kp"`
			Imdb float32 `json:"imdb"`
		}
		Year        int    `json:"year"`
		Description string `json:"description"`
		Poster      struct {
			Url string `json:"url"`
		}
		AltName string `json:"alternativeName"`
	}
	if err = json.NewDecoder(strings.NewReader(string(resp))).Decode(&movieInfo); err != nil {
		return domain.Movie{}, err
	}

	return domain.Movie{
		ID:      movieInfo.ID,
		Name:    movieInfo.Name,
		EngName: movieInfo.AltName,
		Poster:  movieInfo.Poster.Url,
		Rating:  movieInfo.Rating.Kp,
		Year:    movieInfo.Year,
	}, nil

}

func (repo *Repo) GetMoviesIDByActorID(actorID int) ([]int, error) {

	req := fmt.Sprintf("person/%d", actorID)

	resp, err := repo.doRequest(req)
	if err != nil {
		return nil, err
	}

	var actorInfo struct {
		ID     int `json:"id"`
		Movies []struct {
			Id         int    `json:"id"`
			Profession string `json:"enProfession"`
		} `json:"movies"`
	}
	if err = json.NewDecoder(strings.NewReader(string(resp))).Decode(&actorInfo); err != nil {
		return nil, err
	}

	result := make([]int, 0, len(actorInfo.Movies))
	for _, movie := range actorInfo.Movies {
		if movie.Profession == "actor" {
			result = append(result, movie.Id)
		}
	}

	return result, nil

}

func (repo *Repo) SearchActors(query string) ([]domain.Actor, error) {
	encodedQuery := url.QueryEscape(query)
	req := fmt.Sprintf("person/search?page=1&limit=20&query=%s", encodedQuery)

	resp, err := repo.doRequest(req)
	if err != nil {
		return nil, err
	}

	var actors struct {
		Docs []domain.Actor `json:"docs"`
	}
	if err = json.NewDecoder(strings.NewReader(string(resp))).Decode(&actors); err != nil {
		return nil, err
	}

	filtered := filteringActors(actors.Docs)
	result := make([]domain.Actor, 0, 3)
	if len(filtered) >= 3 {
		result = append(result, filtered[:3]...)
	} else {
		result = append(result, filtered...)
	}
	return result, nil
}

func (repo *Repo) doRequest(endpoint string) ([]byte, error) {
	const op = "Repo.doRequest"
	req, err := http.NewRequest("GET", repo.Path+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create request:%w", op, err)
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("X-API-KEY", repo.APIKey)

	resp, err := repo.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: request failed: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s: bad status %d, response: %s", op, resp.StatusCode, body)
	}

	return io.ReadAll(resp.Body)
}

func filteringActors(actors []domain.Actor) []domain.Actor {
	filtered := make([]domain.Actor, 0)

	for _, actor := range actors {
		if actor.Photo != "" && actor.Name != "" {
			filtered = append(filtered, actor)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return len(filtered[i].Movies) < len(filtered[j].Movies)
	})

	return filtered
}
