package pubsub

import (
	"sync"

	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/satori/go.uuid"
)

type Hub interface {
	Subscribe() *Subscription
	Publish(event *Event)
	Size() int
}

type CLAChan chan *cp.ClusterLoadAssignment
type EventChan chan *Event

type Event struct {
	CLA      *cp.ClusterLoadAssignment
	Clusters []*cp.Cluster
	Routes   []*cp.RouteConfiguration
}

type hub struct {
	subscriptions sync.Map
}

func (h *hub) Subscribe() *Subscription {
	id := uuid.NewV4()
	subs := &Subscription{ID: id, Events: make(EventChan, 1000), OnClose: func(subID uuid.UUID) {
		h.subscriptions.Delete(subID)
	}}
	h.subscriptions.Store(id, subs)

	return subs
}

func (h *hub) Publish(event *Event) {
	h.subscriptions.Range(func(id, subscription interface{}) bool {
		subscription.(*Subscription).Accept(event)
		return true
	})
}

func (h *hub) Size() int {
	var ids []interface{}
	h.subscriptions.Range(func(id, subscription interface{}) bool {
		ids = append(ids, id)
		return true
	})
	return len(ids)
}

func NewHub() Hub {
	return &hub{}
}
