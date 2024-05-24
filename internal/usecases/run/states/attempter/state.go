package attempter

import (
	"context"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/config"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/depgraph/factory"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/go-zookeeper/zk"
	"log/slog"
	"sort"
	"strings"
	"time"
)

const electionPath = "/election"

func New(logger *slog.Logger, config config.Config, conn *zk.Conn, factory factory.StateFactory) *State {
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
	return "Attempter"
}

func (s *State) Run(ctx context.Context) (states.AutomataState, error) {

	s.logger.LogAttrs(ctx, slog.LevelInfo, "Attempting to become leader")

	znode, err := s.conn.CreateProtectedEphemeralSequential(electionPath+"/guid-n_", nil, zk.WorldACL(zk.PermAll))

	if err != nil {
		s.logger.LogAttrs(ctx, slog.LevelError, "Error creating znode", slog.String("error", err.Error()))
		return s.factory.GetFailoverState()
	}
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Created znode", slog.String("znode", znode))

	for {
		select {
		case <-ctx.Done():
			return s.factory.GetStoppingState()
		case <-time.After(time.Second * 5):

			children, _, err := s.conn.Children(electionPath)
			if err != nil {
				s.logger.LogAttrs(ctx, slog.LevelError, "Error getting children", slog.String("error", err.Error()))
				continue
			}
			sort.SliceStable(children, func(i, j int) bool {
				znodeSeqI := strings.Split(children[i], "guid-n_")[1]
				znodeSeqJ := strings.Split(children[j], "guid-n_")[1]
				return znodeSeqI < znodeSeqJ
			})

			if znode == electionPath+"/"+children[0] {
				s.logger.LogAttrs(ctx, slog.LevelInfo, "I am the leader")
				return s.factory.GetLeaderState()
			}

			// Find the znode to watch
			index := sort.SearchStrings(children, znode[len(electionPath)+1:])
			if index > 0 {
				previousZnode := children[index-1]
				_, _, ch, err := s.conn.ExistsW(electionPath + "/" + previousZnode)
				if err != nil {
					s.logger.LogAttrs(ctx, slog.LevelError, "Error setting watch", slog.String("error", err.Error()))
					continue
				}
				s.logger.LogAttrs(ctx, slog.LevelInfo, "Watching znode", slog.String("znode", previousZnode))
				<-ch
			}
		}
	}
}
