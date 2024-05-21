package run

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
)

var _ Runner = &LoopRunner{}

type Runner interface {
	Run(ctx context.Context, state states.AutomataState) error
}

func NewLoopRunner(logger *slog.Logger, leaderTimeout, attempterTimeout time.Duration, fileDir string, storageCapacity int) *LoopRunner {
	logger = logger.With("subsystem", "StateRunner")
	return &LoopRunner{
		logger:           logger,
		leaderTimeout:    leaderTimeout,
		attempterTimeout: attempterTimeout,
		fileDir:          fileDir,
		storageCapacity:  storageCapacity,
	}
}

type LoopRunner struct {
	logger           *slog.Logger
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	leaderTimeout    time.Duration
	attempterTimeout time.Duration
	fileDir          string
	storageCapacity  int
}

func (r *LoopRunner) Run(ctx context.Context, state states.AutomataState) error {
	r.ctx, r.cancel = context.WithCancel(ctx)
	defer r.cancel()
	defer r.wg.Wait()

	for state != nil {
		select {
		case <-r.ctx.Done():
			return fmt.Errorf("state machine stopped: %w", r.ctx.Err())
		default:
			r.logger.LogAttrs(r.ctx, slog.LevelInfo, "start running state", slog.String("state", state.String()))
			var err error
			state, err = state.Run(r.ctx)
			if err != nil {
				return fmt.Errorf("state %s run: %w", state.String(), err)
			}
		}
	}
	r.logger.LogAttrs(r.ctx, slog.LevelInfo, "no new state, finish")
	return nil
}
