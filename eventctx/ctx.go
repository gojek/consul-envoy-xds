package eventctx

import (
	"context"
	"github.com/newrelic/go-agent"
	"net/http"
)

const newRelicTxnKey = "CTX_NEW_RELIC_TXN_KEY"

type noOpTransaction struct {
	http.ResponseWriter
}

func (t noOpTransaction) End() error {
	return nil
}

func (t noOpTransaction) Ignore() error {
	return nil
}

func (t noOpTransaction) SetName(name string) error {
	return nil
}

func (t noOpTransaction) NoticeError(err error) error {
	return nil
}

func (t noOpTransaction) AddAttribute(key string, value interface{}) error {
	return nil
}

func (t noOpTransaction) StartSegmentNow() newrelic.SegmentStartTime {
	return newrelic.SegmentStartTime{}
}

func NewRelicTxn(ctx context.Context) newrelic.Transaction {
	val := ctx.Value(newRelicTxnKey)
	if val == nil {
		return noOpTransaction{}
	}
	return val.(newrelic.Transaction)
}

func SetNewRelicTxn(ctx context.Context, txn newrelic.Transaction) context.Context {
	return context.WithValue(ctx, newRelicTxnKey, txn)
}
