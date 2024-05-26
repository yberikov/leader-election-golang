package initial

import (
	"context"
	"log/slog"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/config"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/depgraph/factory"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/go-zookeeper/zk"
)

const electionPath = "/election"

func New(logger *slog.Logger, config config.Config, conn *zk.Conn, factory factory.StateFactory) *State {
	logger = logger.With("state", "InitState")
	return &State{
		logger:  logger,
		conn:    conn,
		config:  config,
		factory: factory,
	}
}

type State struct {
	logger  *slog.Logger
	conn    *zk.Conn
	config  config.Config
	factory factory.StateFactory
}

// String returns the name of the state
func (s *State) String() string {
	return "InitState"
}

// Run executes the logic of the Init state
func (s *State) Run(ctx context.Context) (states.AutomataState, error) {
	if s.conn == nil {
		conn, _, err := zk.Connect(s.config.ZookeeperServers, 10*time.Second)
		s.conn = conn
		if err != nil {
			s.logger.Error("Connection failed in initState", "error", err)
			return s.factory.GetFailoverState()
		}
	}
	select {
	case <-ctx.Done():
		s.logger.LogAttrs(ctx, slog.LevelInfo, "Context done in init state")
		return s.factory.GetStoppingState()
	default:
		// Ensure the election znode exists
		exists, _, err := s.conn.Exists(electionPath)
		if err != nil {
			s.logger.Error("Error checking if znode exists:", "error", err)
			return s.factory.GetFailoverState()
		}
		if !exists {
			_, err := s.conn.Create(electionPath, nil, 0, zk.WorldACL(zk.PermAll))
			if err != nil {
				s.logger.Error("Error creating election znode", "error", err)
				return s.factory.GetFailoverState()
			}
		}
		err = s.factory.SetConn(s.conn)
		if err != nil {
			s.logger.Error("Error on setting connection", "error", err)
			return s.factory.GetFailoverState()
		}
		return s.factory.GetAttempterState()
	}
}
