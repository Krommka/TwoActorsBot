package SessionStates

import (
	"KinopoiskTwoActors/internal/domain"
	"context"
	"github.com/google/uuid"
	"sync"
)

type SessionStates struct {
	states map[int64]*domain.SessionState
	mu     sync.RWMutex
}

func NewUserStates() *SessionStates {
	states := make(map[int64]*domain.SessionState)
	return &SessionStates{
		states: states,
		mu:     sync.RWMutex{},
	}
}

func (s *SessionStates) GetCurrentStatesID(ctx context.Context) []int64 {
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

func (s *SessionStates) GetStateByID(ctx context.Context, chatID int64) *domain.SessionState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.states[chatID]; !ok {
		s.states[chatID] = &domain.SessionState{
			SentMediaMessages: []int{},
			TempActors:        []domain.PhotoData{},
		}
	}
	return s.states[chatID]
}

func (s *SessionStates) ResetUserState(ctx context.Context, chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.states, chatID)
}

func (s *SessionStates) GetCorrelationID(ctx context.Context, chatID int64) string {
	state := s.GetStateByID(ctx, chatID)
	if state.CorrelationID == "" {
		state.CorrelationID = generateCorrelationID()
	}
	return state.CorrelationID
}

func generateCorrelationID() string {
	return uuid.New().String()
}

func (s *SessionStates) SetState(ctx context.Context, chatID int64, state *domain.SessionState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[chatID] = state
	return nil
}
