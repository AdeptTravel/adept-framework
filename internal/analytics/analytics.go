package analytics

import (
	"context"
	"time"
)

type Writer struct {
	Add func(ctx context.Context, topic string, evt any) error
}

type Event struct {
	Name string            `json:"name"`
	Tags map[string]string `json:"tags"`
	Ts   time.Time         `json:"ts"`
	Val  float64           `json:"val"`
}

func (w Writer) Count(ctx context.Context, name string, tags map[string]string) {
	_ = w.Add(ctx, "analytics", Event{
		Name: name, Tags: tags, Ts: time.Now(), Val: 1,
	})
}
