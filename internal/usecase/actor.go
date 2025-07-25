package usecase

import (
	"KinopoiskTwoActors/internal/domain"
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type Actor struct {
	repo ActorFilmRepository
}

func NewActor(repo ActorFilmRepository) *Actor {
	return &Actor{repo: repo}
}

func (uc *Actor) SearchActor(ctx context.Context, query string) ([]domain.Actor, error) {
	const op = "useCase.ActorSearcher"

	if len(query) == 0 {
		return nil, fmt.Errorf("%s:empty query", op)
	}

	actors, err := uc.repo.SearchActors(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("%s:repo error: %v", op, err)
	}

	if len(actors) == 0 {
		return nil, fmt.Errorf("%s:actors not found", op)
	}

	normalizedQuery := normalizeName(query)
	sameNameActors := make([]domain.Actor, 0)
	for _, actor := range actors {
		if (normalizeName(actor.Name) == normalizedQuery && actor.PhotoURL != "") ||
			(normalizeName(actor.EngName) == normalizedQuery && actor.PhotoURL != "") {
			sameNameActors = append(sameNameActors, actor)
			break
		}
	}
	if len(sameNameActors) > 0 {
		return sameNameActors, nil
	}

	filtered := filteringActors(actors)
	result := make([]domain.Actor, 0, 3)
	if len(filtered) >= 3 {
		result = append(result, filtered[:3]...)
	} else {
		result = append(result, filtered...)
	}
	return result, nil
}

func normalizeName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	reg := regexp.MustCompile(`[^a-zа-яё]`)
	return reg.ReplaceAllString(name, "")
}

func filteringActors(actors []domain.Actor) []domain.Actor {
	filtered := make([]domain.Actor, 0)

	for _, actor := range actors {
		if actor.PhotoURL != "" && actor.Name != "" {
			filtered = append(filtered, actor)
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		return len(filtered[i].Movies) < len(filtered[j].Movies)
	})

	return filtered
}
