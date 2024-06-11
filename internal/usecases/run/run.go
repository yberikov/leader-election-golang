package run

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/depgraph/factory"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
)

var _ Runner = &LoopRunner{}

type Runner interface {
	Run(ctx context.Context, state states.AutomataState) error
}

func NewLoopRunner(logger *slog.Logger, factory factory.StateFactory) *LoopRunner {
	logger = logger.With("subsystem", "StateRunner")
	return &LoopRunner{
		logger:  logger,
		factory: factory,
	}
}

type LoopRunner struct {
	logger  *slog.Logger
	factory factory.StateFactory
}

func (r *LoopRunner) Run(ctx context.Context, state states.AutomataState) error {
	for state != nil {
		select {
		case <-ctx.Done():
			r.logger.LogAttrs(ctx, slog.LevelInfo, "Context cancelled, transitioning to stopping state")
			stoppingState, _ := r.factory.GetStoppingState()
			_, err := stoppingState.Run(ctx)
			return err
		default:
			r.logger.LogAttrs(ctx, slog.LevelInfo, "start running state", slog.String("state", state.String()))
			var err error
			state, err = state.Run(ctx)
			if err != nil {
				return fmt.Errorf("state %s run: %w", state.String(), err)
			}
		}
	}
	r.logger.LogAttrs(ctx, slog.LevelInfo, "no new state, finish")
	return nil
}
