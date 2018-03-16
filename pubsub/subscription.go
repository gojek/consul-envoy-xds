package pubsub

import (
	"github.com/satori/go.uuid"
)

//Subscription is a unique channel of ClusterLoadAssignments
type Subscription struct {
	ID      uuid.UUID
	Events  EventChan
	OnClose func(uuid.UUID)
}

func (s Subscription) Accept(e *Event) {
	select {
	case s.Events <- e:
	}
}

func (s Subscription) Close() {
	s.OnClose(s.ID)
	close(s.Events)
}
