package bus

import (
	"context"
	"reflect"
	"sync"
)

type handler struct {
	typ  reflect.Type
	fn   reflect.Value
}

type inMem struct {
	sync.RWMutex
	subs map[string][]handler
}

func New() *inMem { return &inMem{subs: make(map[string][]handler)} }

// Publish fan-outs a value to all handlers for that topic.
func (b *inMem) Publish(ctx context.Context, topic string, v any) {
	b.RLock()
	list := append([]handler(nil), b.subs[topic]...)
	b.RUnlock()

	for _, h := range list {
		if reflect.TypeOf(v) != h.typ { continue }
		go h.fn.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(v)})
	}
}

// Subscribe registers fn(ctx, T) for a generic T.
func (b *inMem) Subscribe[T any](topic string, fn func(context.Context, T)) {
	param := reflect.TypeOf(fn).In(1)
	b.Lock()
	b.subs[topic] = append(b.subs[topic], handler{
		typ: param, fn: reflect.ValueOf(fn),
	})
	b.Unlock()
}