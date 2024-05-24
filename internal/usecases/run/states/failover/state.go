package failover

import (
	"context"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/config"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/depgraph/factory"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/go-zookeeper/zk"
	"log/slog"
)

// New creates a new instance of the Failover state
func New(logger *slog.Logger, config config.Config, factory factory.StateFactory) *State {
	logger = logger.With("subsystem", "FailoverState")
	return &State{
		logger:  logger,
		config:  config,
		factory: factory,
	}
}

// State represents the Failover state of the state machine
type State struct {
	logger  *slog.Logger
	config  config.Config
	factory factory.StateFactory
}

// String returns the name of the state
func (s *State) String() string {
	return "Failover"
}

// Run executes the logic of the Failover state
func (s *State) Run(ctx context.Context) (states.AutomataState, error) {
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Entering failover state")

	retryIntervals := []time.Duration{0, 1 * time.Second, 2 * time.Second, 5 * time.Second, 10 * time.Second}

	for i, interval := range retryIntervals {
		select {
		case <-ctx.Done():
			return s.factory.GetStoppingState()
		case <-time.After(interval):
			s.logger.LogAttrs(ctx, slog.LevelInfo, "Attempting to recover connection to Zookeeper", slog.Int("attempt", i+1))

			conn, _, err := zk.Connect(s.config.ZookeeperServers, 10*time.Second)
			if conn != nil && err != nil {
				s.logger.LogAttrs(ctx, slog.LevelInfo, "Successfully reconnected to Zookeeper")
				// Assuming that the Init state is the entry point after a successful reconnection
				initState, err := s.factory.GetInitState()
				if err != nil {
					s.logger.LogAttrs(ctx, slog.LevelError, "Failed to get init state after recovery", slog.String("error", err.Error()))
					return s.factory.GetStoppingState()
				}
				return initState, nil
			}
			s.logger.LogAttrs(ctx, slog.LevelError, "Failed to reconnect to Zookeeper", slog.String("error", err.Error()))
		}
	}

	s.logger.LogAttrs(ctx, slog.LevelError, "Max recovery attempts reached, transitioning to stopping state")
	return s.factory.GetStoppingState()
}
