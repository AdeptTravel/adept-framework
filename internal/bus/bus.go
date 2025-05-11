package bus

import (
	"context"
	"reflect"
	"sync"
)

type handler struct {
	typ reflect.Type
	fn  reflect.Value
}

// internal/bus/bus.go

// Bus is the public behaviour every module relies on.
type Bus interface {
	Publish(ctx context.Context, topic string, v any)
	Subscribe(topic string, fn interface{})
}

// InMem is the simple in-process implementation.
type InMem struct { // renamed (capital I)
	sync.RWMutex
	subs map[string][]handler
}

func New() *InMem { return &InMem{subs: make(map[string][]handler)} }

// Publish fan-outs a value to all handlers for that topic.
func (b *InMem) Publish(ctx context.Context, topic string, v any) {
	b.RLock()
	list := append([]handler(nil), b.subs[topic]...)
	b.RUnlock()

	for _, h := range list {

		if reflect.TypeOf(v) != h.typ {
			continue
		}
		h.fn.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(v)})

	}
}

// Subscribe registers fn(ctx, T) for a generic T.

func (b *InMem) Subscribe(topic string, fn interface{}) {
	b.Lock()
	h := handler{
		typ: reflect.TypeOf(fn).In(1), // second param's type
		fn:  reflect.ValueOf(fn),
	}
	b.subs[topic] = append(b.subs[topic], h)
	b.Unlock()
}
