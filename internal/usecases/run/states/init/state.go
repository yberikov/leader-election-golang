package initial

import (
	"context"
	"errors"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/attempter"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/failover"
	"github.com/go-zookeeper/zk"
	"log/slog"
	"time"
)

// New creates a new instance of the Init state
func New(logger *slog.Logger, zooServers []string, attemptInterval time.Duration, writeInterval time.Duration) *State {
	return &State{
		logger:          logger,
		zooServers:      zooServers,
		attemptInterval: attemptInterval,
		writeInterval:   writeInterval,
	}
}

// State represents the Init state of the state machine
type State struct {
	logger          *slog.Logger
	zooServers      []string
	attemptInterval time.Duration
	writeInterval   time.Duration
}

// String returns the name of the state
func (s *State) String() string {
	return "InitState"
}

// Run executes the logic of the Init state
func (s *State) Run(ctx context.Context) (states.AutomataState, error) {
	conn, _, err := zk.Connect(s.zooServers, 10*time.Second)
	if err != nil {
		s.logger.Error("Connection failed")
		return failover.New(s.logger), errors.New("connection to zookeeper failed")
	}
	// Define the election path
	electionPath := "/election"

	// Ensure the election znode exists
	exists, _, err := conn.Exists(electionPath)
	if err != nil {
		s.logger.Error("Error checking if znode exists: %v", err)
		return failover.New(s.logger), errors.New("connection to zookeeper failed")
	}
	if !exists {
		_, err := conn.Create(electionPath, nil, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			s.logger.Error("Error creating election znode: %v", err)
			return failover.New(s.logger), errors.New("connection to zookeeper failed")
		}
	}

	return attempter.New(s.logger, conn, s.attemptInterval, s.writeInterval), nil
}
