package factory

import (
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/go-zookeeper/zk"
)

type StateFactory interface {
	GetInitState() (states.AutomataState, error)
	GetFailoverState() (states.AutomataState, error)
	GetAttempterState() (states.AutomataState, error)
	GetLeaderState() (states.AutomataState, error)
	GetStoppingState() (states.AutomataState, error)
	SetConn(conn *zk.Conn) error
}
