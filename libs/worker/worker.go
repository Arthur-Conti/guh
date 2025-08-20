package worker

import (
	"context"
	"time"

	fl "github.com/Arthur-Conti/guh/libs/fast_logger"
	httphandler "github.com/Arthur-Conti/guh/libs/http_handler"
	"github.com/google/uuid"
)

type Worker[T any] struct {
	ID       uuid.UUID
	channel  chan any
	Response T
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

func (w *Worker[T]) DoAndNotify(function func() (T, error), url string) {
	go func() {
		response, err := function()
		if err != nil {
			httphandler.NewHttpHandler().Request("POST", url, map[string]any{
				"error": err.Error(),
			}, nil)
		}
		err = httphandler.NewHttpHandler().Request("POST", url, response, nil)
		if err != nil {
			fl.Logf("Error sending response to %s: %v", url, err)
		}
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
