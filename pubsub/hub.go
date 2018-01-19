package pubsub

import (
	"sync"

	cp "github.com/envoyproxy/go-control-plane/api"
	"github.com/satori/go.uuid"
)

type Hub interface {
	Subscribe() *Subscription
	Publish(assignment *cp.ClusterLoadAssignment)
	Size() int
}

type CLAChan chan *cp.ClusterLoadAssignment

type hub struct {
	subscriptions sync.Map
}

func (h *hub) Subscribe() *Subscription {
	id := uuid.NewV4()
	subs := &Subscription{ID: id, Cla: make(CLAChan, 1000), OnClose: func(subID uuid.UUID) {
		h.subscriptions.Delete(subID)
	}}
	h.subscriptions.Store(id, subs)

	return subs
}

func (h *hub) Publish(assignment *cp.ClusterLoadAssignment) {
	h.subscriptions.Range(func(id, subscription interface{}) bool {
		subscription.(*Subscription).Accept(assignment)
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
