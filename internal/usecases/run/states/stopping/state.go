package stopping

import (
	"context"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"log/slog"
)

// New creates a new instance of the Init state
func New(logger *slog.Logger) *State {
	logger = logger.With("subsystem", "StoppingState")
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
	return "Stopping"
}

// Run executes the logic of the Init state
func (s *State) Run(ctx context.Context) (states.AutomataState, error) {
	//TODO release all resources

	select {}
}
