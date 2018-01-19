package pubsub

import (
	cp "github.com/envoyproxy/go-control-plane/api"
	"github.com/satori/go.uuid"
)

//Subscription is a unique channel of ClusterLoadAssignments
type Subscription struct {
	ID      uuid.UUID
	Cla     CLAChan
	OnClose func(uuid.UUID)
}

func (s Subscription) Accept(c *cp.ClusterLoadAssignment) {
	select {
	case s.Cla <- c:
	}
}

func (s Subscription) Close() {
	s.OnClose(s.ID)
	close(s.Cla)
}
