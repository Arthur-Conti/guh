package retryhandler

import (
	"time"

	errorhandler "github.com/Arthur-Conti/guh/libs/error_handler"
)

type RetryOpts struct {
	MaxAttempts int
	Backoff     int
}

type RetryHandler struct {
	opts RetryOpts
}

func NewRetryHandler(opts RetryOpts) *RetryHandler {
	return &RetryHandler{opts: opts}
}

func (rh *RetryHandler) Do(function func() error, opts RetryOpts, df bool) error {
	var maxAttempts int
	var backoff int
	if df {
		maxAttempts = rh.opts.MaxAttempts
		backoff = rh.opts.Backoff
	} else {
		maxAttempts = opts.MaxAttempts
		backoff = opts.Backoff
	}
	err := errorhandler.New(errorhandler.BadGateway, "starting error") 
	var count int
	for err != nil && count < maxAttempts {
		err = function()
		count += 1
		if backoff != 0 {
			time.Sleep(time.Duration(backoff)*time.Second)
		}
	}
	return nil
}
