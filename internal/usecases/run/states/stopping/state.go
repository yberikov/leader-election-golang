package stopping

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/config"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/depgraph/factory"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/go-zookeeper/zk"
)

func New(logger *slog.Logger, conn *zk.Conn, config config.Config, factory factory.StateFactory) *State {
	logger = logger.With("state", "StoppingState")
	return &State{
		logger:  logger,
		conn:    conn,
		config:  config,
		factory: factory,
	}
}

// State represents the Init state of the state machine
type State struct {
	logger  *slog.Logger
	conn    *zk.Conn
	config  config.Config
	factory factory.StateFactory
}

// String returns the name of the state
func (s *State) String() string {
	return "Stopping"
}

// Run executes the logic of the Init state
func (s *State) Run(ctx context.Context) (states.AutomataState, error) {
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Entering stopping state")

	s.logger.LogAttrs(ctx, slog.LevelInfo, "Releasing resources")
	if s.conn != nil {
		s.conn.Close()
	}

	s.logger.LogAttrs(ctx, slog.LevelInfo, "Application stopped gracefully")
	return nil, fmt.Errorf("stopping state: application stopped gracefully")
}
