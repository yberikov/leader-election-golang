package attempter

import (
	"context"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/failover"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/leader"
	"github.com/go-zookeeper/zk"
	"log/slog"
	"sort"
	"strings"
	"time"
)

const electionPath = "/election"

// New creates a new instance of the Init state
func New(logger *slog.Logger, conn *zk.Conn, attemptInterval time.Duration, writeInterval time.Duration) *State {
	return &State{
		logger:          logger,
		conn:            conn,
		attemptInterval: attemptInterval,
		writeInterval:   writeInterval,
	}
}

// State represents the Init state of the state machine
type State struct {
	logger          *slog.Logger
	conn            *zk.Conn
	attemptInterval time.Duration
	writeInterval   time.Duration
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
		return failover.New(s.logger), nil
	}
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Created znode", slog.String("znode", znode))

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
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
				return leader.New(s.logger, s.writeInterval), nil
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
