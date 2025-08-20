package worker

import (
	"context"
	"time"

	httphandler "github.com/Arthur-Conti/guh/libs/http_handler"
	"github.com/google/uuid"
)

type Worker[T any] struct {
	ID       uuid.UUID
	channel  chan any
	response T
}

func NewWorker[T any](id uuid.UUID) *Worker[T] {
	return &Worker[T]{
		ID:      id,
		channel: make(chan any),
	}
}

func (w *Worker[T]) Do(function func()) {
	go function()
}

func (w *Worker[T]) DoAndNotify(function func() error, url string) {
	go func() {
		err := function()
		if err != nil {
			httphandler.NewHttpHandler().Request("POST", url, map[string]any{
				"error": err.Error(),
			}, nil)
		}
		httphandler.NewHttpHandler().Request("POST", url, w.response, nil)
	}()
}

func (w *Worker[T]) DoWithContext(ctx context.Context, function func()) {
	go function()
}

func (w *Worker[T]) Start(function func(channel chan any)) {
	go function(w.channel)
}

func (w *Worker[T]) Stop() {
	close(w.channel)
}

func (w *Worker[T]) StopAfter(duration time.Duration) {
	time.Sleep(duration)
	w.Stop()
}

func (w *Worker[T]) Pool(functions []func()) {
	for _, function := range functions {
		go function()
	}
}
