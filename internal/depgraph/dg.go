package depgraph

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/config"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/attempter"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/failover"
	initial2 "github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/init"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/leader"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/stopping"
	"github.com/go-zookeeper/zk"
)

type dgEntity[T any] struct {
	sync.Once
	value   T
	initErr error
}

func (e *dgEntity[T]) get(init func() (T, error)) (T, error) {
	e.Do(func() {
		e.value, e.initErr = init()
	})

	if e.initErr != nil {
		return *new(T), e.initErr
	}
	return e.value, nil
}

type DepGraph struct {
	Config         config.Config
	logger         *dgEntity[*slog.Logger]
	stateRunner    *dgEntity[*run.LoopRunner]
	emptyState     *dgEntity[states.AutomataState]
	initState      *dgEntity[states.AutomataState]
	attempterState *dgEntity[states.AutomataState]
	leaderState    *dgEntity[states.AutomataState]
	failoverState  *dgEntity[states.AutomataState]
	stoppingState  *dgEntity[states.AutomataState]
	conn           *zk.Conn
}

func New(config config.Config) *DepGraph {
	return &DepGraph{
		Config:         config,
		logger:         &dgEntity[*slog.Logger]{},
		stateRunner:    &dgEntity[*run.LoopRunner]{},
		emptyState:     &dgEntity[states.AutomataState]{},
		initState:      &dgEntity[states.AutomataState]{},
		attempterState: &dgEntity[states.AutomataState]{},
		leaderState:    &dgEntity[states.AutomataState]{},
		failoverState:  &dgEntity[states.AutomataState]{},
		stoppingState:  &dgEntity[states.AutomataState]{},
	}
}

func (dg *DepGraph) GetLogger() (*slog.Logger, error) {
	return dg.logger.get(func() (*slog.Logger, error) {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})), nil
	})
}

func (dg *DepGraph) GetInitState() (states.AutomataState, error) {
	return dg.initState.get(func() (states.AutomataState, error) {
		logger, err := dg.GetLogger()
		if err != nil {
			return nil, fmt.Errorf("error on: getting logger - %w", err)
		}
		return initial2.New(logger, dg.Config, nil, dg), nil
	})
}

func (dg *DepGraph) GetAttempterState() (states.AutomataState, error) {
	return dg.attempterState.get(func() (states.AutomataState, error) {
		logger, err := dg.GetLogger()
		if err != nil {
			return nil, fmt.Errorf("error on: getting logger - %w", err)
		}
		if dg.conn == nil {
			return nil, fmt.Errorf("error on: Zookeeper connection not established")
		}
		return attempter.New(logger, dg.Config, dg.conn, dg), nil
	})
}

func (dg *DepGraph) GetLeaderState() (states.AutomataState, error) {
	return dg.leaderState.get(func() (states.AutomataState, error) {
		logger, err := dg.GetLogger()
		if err != nil {
			return nil, fmt.Errorf("error on: getting logger %w", err)
		}
		return leader.New(logger, dg.Config, dg), nil
	})
}

func (dg *DepGraph) GetFailoverState() (states.AutomataState, error) {
	return dg.failoverState.get(func() (states.AutomataState, error) {
		logger, err := dg.GetLogger()
		if err != nil {
			return nil, fmt.Errorf("error on: getting logger %w", err)
		}
		return failover.New(logger, dg.Config, dg), nil
	})
}

func (dg *DepGraph) GetStoppingState() (states.AutomataState, error) {
	return dg.stoppingState.get(func() (states.AutomataState, error) {
		logger, err := dg.GetLogger()
		if err != nil {
			return nil, fmt.Errorf("error on: getting logger %w", err)
		}
		return stopping.New(logger, dg.conn, dg.Config, dg), nil
	})
}

func (dg *DepGraph) GetRunner() (run.Runner, error) {
	return dg.stateRunner.get(func() (*run.LoopRunner, error) {
		logger, err := dg.GetLogger()
		if err != nil {
			return nil, fmt.Errorf("error on: getting logger: %w", err)
		}
		return run.NewLoopRunner(logger, dg), nil
	})
}

func (dg *DepGraph) SetConn(conn *zk.Conn) error {
	dg.conn = conn
	return nil
}
