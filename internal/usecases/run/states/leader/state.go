package leader

import (
	"context"
	"fmt"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/config"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/depgraph/factory"
	"io/ioutil"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
)

// New creates a new instance of the Leader state
func New(logger *slog.Logger, config config.Config, factory factory.StateFactory) *State {
	logger = logger.With("subsystem", "LeaderState")
	return &State{
		logger:  logger,
		config:  config,
		factory: factory,
	}
}

// State represents the Leader state of the state machine
type State struct {
	logger  *slog.Logger
	config  config.Config
	factory factory.StateFactory
}

func (s *State) String() string {
	return "Leader"
}

func (s *State) Run(ctx context.Context) (states.AutomataState, error) {
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Became leader, starting work")

	ticker := time.NewTicker(s.config.LeaderTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.LogAttrs(ctx, slog.LevelInfo, "Context done in leader state")
			return s.factory.GetStoppingState()
		case <-ticker.C:
			filePath := filepath.Join(s.config.FileDir, fmt.Sprintf("leader_%d.txt", time.Now().Unix()))
			err := os.WriteFile(filePath, []byte("Leader active"), 0644)
			if err != nil {
				s.logger.LogAttrs(ctx, slog.LevelError, "Error writing to file", slog.String("error", err.Error()))
				continue
			}
			s.logger.LogAttrs(ctx, slog.LevelInfo, "Wrote to file", slog.String("file", filePath))

			// Manage files in the directory
			err = s.manageFiles(ctx)
			if err != nil {
				s.logger.LogAttrs(ctx, slog.LevelError, "Error managing files", slog.String("error", err.Error()))
			}
		}
	}
}

// manageFiles ensures that the number of files in the directory does not exceed the storage capacity
func (s *State) manageFiles(ctx context.Context) error {
	files, err := ioutil.ReadDir(s.config.FileDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	if len(files) > s.config.StorageCapacity {
		// Sort files by modification time
		sort.Slice(files, func(i, j int) bool {
			return files[i].ModTime().Before(files[j].ModTime())
		})

		// Remove oldest files if exceeding storage capacity
		excess := len(files) - s.config.StorageCapacity
		for i := 0; i < excess; i++ {
			filePath := filepath.Join(s.config.FileDir, files[i].Name())
			err := os.Remove(filePath)
			if err != nil {
				return fmt.Errorf("failed to remove file %s: %w", filePath, err)
			}
			s.logger.LogAttrs(ctx, slog.LevelInfo, "Removed file", slog.String("file", filePath))
		}
	}

	return nil
}
