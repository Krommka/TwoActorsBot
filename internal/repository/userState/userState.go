package userState

import (
	"KinopoiskTwoActors/internal/domain"
	"sync"
)

type State struct {
	CorrelationID     string
	Step              string
	FirstActorID      int
	SecondActorID     int
	SentMediaMessages []int
	TempActors        []domain.PhotoData
}

type UserStates struct {
	states map[int64]*State
	mu     sync.RWMutex
}

func NewUserStates() *UserStates {
	states := make(map[int64]*State)
	return &UserStates{
		states: states,
		mu:     sync.RWMutex{},
	}
}

func (s *UserStates) GetCurrentStatesID() []int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	states := make([]int64, 0, 32)

	for k, v := range s.states {
		if v != nil {
			states = append(states, k)
		}
	}
	return states
}

func (s *UserStates) GetByID(chatID int64) *State {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.states[chatID]; !ok {
		s.states[chatID] = &State{}
	}
	return s.states[chatID]
}

func (s *UserStates) ResetUserState(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.states, chatID)
}
