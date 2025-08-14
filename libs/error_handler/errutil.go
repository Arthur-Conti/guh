package errorhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
)

type Kind int

const (
	KindUnknown Kind = iota
	KindInvalidArgument
	KindUnauthenticated
	KindPermissionDenied
	KindNotFound
	KindAlreadyExists
	KindResourceExhausted
	KindFailedPrecondition
	KindAborted
	KindOutOfRange
	KindUnimplemented
	KindInternal
	KindUnavailable
	KindDeadlineExceeded
)

func (k Kind) HttpStatus() int {
	switch k {
	case KindInvalidArgument:
		return http.StatusBadRequest
	case KindUnauthenticated:
		return http.StatusUnauthorized
	case KindPermissionDenied:
		return http.StatusForbidden
	case KindNotFound:
		return http.StatusNotFound
	case KindAlreadyExists:
		return http.StatusConflict
	case KindResourceExhausted:
		return http.StatusTooManyRequests
	case KindFailedPrecondition:
		return http.StatusPreconditionFailed
	case KindAborted:
		return http.StatusConflict
	case KindOutOfRange:
		return http.StatusRequestedRangeNotSatisfiable
	case KindUnimplemented:
		return http.StatusNotImplemented
	case KindInternal:
		return http.StatusInternalServerError
	case KindUnavailable:
		return http.StatusServiceUnavailable
	case KindDeadlineExceeded:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

type Error struct {
	Kind    Kind
	Op      string
	Message string
	Fields  map[string]any
	Cause   error
	stack   []uintptr
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%v] %v: %v", e.Op, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%v] %v", e.Op, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Cause
}

type Option func(*Error)

func WithOp(op string) Option {
	return func(e *Error) {
		e.Op = op
	}
}

func WithFields(f map[string]any) Option {
	return func(e *Error) {
		e.Fields = f
	}
}

func WithCause(cause error) Option {
	return func(e *Error) {
		e.Cause = cause
	}
}

func WithStack() Option {
	return func(e *Error) {
		e.stack = make([]uintptr, 10)
		e.stack = e.stack[:runtime.Callers(3, e.stack)]
	}
}

func New(kind Kind, message string, opts ...Option) error {
	e := &Error{Kind: kind, Message: message}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func Wrap(kind Kind, message string, err error, opts ...Option) error {
	e := &Error{Kind: kind, Message: message, Cause: err}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func IsKind(err error, kind Kind) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Kind == kind
	}
	return false
}

func Status(err error) int {
	var e *Error
	if errors.As(err, &e) {
		return e.Kind.HttpStatus()
	}
	return http.StatusInternalServerError
}

type Problem struct {
	Type     string         `json:"type"`
	Title    string         `json:"title"`
	Status   int            `json:"status"`
	Detail   string         `json:"detail"`
	Instance string         `json:"instance"`
	Extras   map[string]any `json:"extras"`
}

func toProblem(err error) (int, []byte) {
	var e *Error
	if !errors.As(err, &e) {
		s := http.StatusInternalServerError
		p := Problem{
			Type:     fmt.Sprintf("about:blank#%d", s),
			Title:    http.StatusText(s),
			Status:   s,
			Detail:   err.Error(),
			Instance: e.Op,
		}
		b, _ := json.Marshal(p)
		return s, b
	}
	s := e.Kind.HttpStatus()
	p := Problem{
		Type:   fmt.Sprintf("about:blank#%d", s),
		Title:  http.StatusText(s),
		Status: s,
		Detail: e.Message,
		Extras: e.Fields,
	}
	b, _ := json.Marshal(p)
	return s, b
}

func Retryable(err error) bool {
	switch {
	case IsKind(err, KindUnavailable), IsKind(err, KindAborted), IsKind(err, KindResourceExhausted):
		return false
	default:
		return false
	}
}
