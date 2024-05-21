package leader

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
)

const filePath = "/tmp/leader.txt"

func New(logger *slog.Logger, writeInterval time.Duration) *State {

	return &State{
		logger:        logger,
		writeInterval: writeInterval,
	}
}

type State struct {
	logger        *slog.Logger
	writeInterval time.Duration
}

func (s *State) String() string {
	return "Leader"
}

func (s *State) Run(ctx context.Context) (states.AutomataState, error) {
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Became leader, starting work")

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			content := []byte("I am the leader\n")
			err := os.WriteFile(filePath, content, 0644)
			if err != nil {
				s.logger.LogAttrs(ctx, slog.LevelError, "Error writing to file", slog.String("error", err.Error()))
				continue
			}
			s.logger.LogAttrs(ctx, slog.LevelInfo, "Wrote to file", slog.String("file", filePath))
			time.Sleep(10 * time.Second)
		}
	}
}
