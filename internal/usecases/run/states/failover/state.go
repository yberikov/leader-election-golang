package failover

import (
	"context"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/stopping"
	"log/slog"
)

// New creates a new instance of the Init state
func New(logger *slog.Logger) *State {
	logger = logger.With("subsystem", "FailoverState")
	return &State{
		logger: logger,
	}
}

// State represents the Init state of the state machine
type State struct {
	logger *slog.Logger
}

// String returns the name of the state
func (s *State) String() string {
	return "Failover"
}

// Run executes the logic of the Init state
func (s *State) Run(ctx context.Context) (states.AutomataState, error) {
	//TODO try again
	return stopping.New(s.logger), nil
}
