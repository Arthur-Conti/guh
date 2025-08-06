package errorhandler

import (
	"errors"
	"fmt"
	"net/http"
)

var statusMap = map[string]int{
	"BadRequest":                   http.StatusBadRequest,                   // 400
	"Unauthorized":                 http.StatusUnauthorized,                 // 401
	"PaymentRequired":              http.StatusPaymentRequired,              // 402
	"Forbidden":                    http.StatusForbidden,                    // 403
	"NotFound":                     http.StatusNotFound,                     // 404
	"MethodNotAllowed":             http.StatusMethodNotAllowed,             // 405
	"NotAcceptable":                http.StatusNotAcceptable,                // 406
	"ProxyAuthRequired":            http.StatusProxyAuthRequired,            // 407
	"RequestTimeout":               http.StatusRequestTimeout,               // 408
	"Conflict":                     http.StatusConflict,                     // 409
	"Gone":                         http.StatusGone,                         // 410
	"LengthRequired":               http.StatusLengthRequired,               // 411
	"PreconditionFailed":           http.StatusPreconditionFailed,           // 412
	"RequestEntityTooLarge":        http.StatusRequestEntityTooLarge,        // 413
	"RequestURITooLong":            http.StatusRequestURITooLong,            // 414
	"UnsupportedMediaType":         http.StatusUnsupportedMediaType,         // 415
	"RequestedRangeNotSatisfiable": http.StatusRequestedRangeNotSatisfiable, // 416
	"ExpectationFailed":            http.StatusExpectationFailed,            // 417
	"Teapot":                       http.StatusTeapot,                       // 418
	"UnprocessableEntity":          http.StatusUnprocessableEntity,          // 422
	"TooManyRequests":              http.StatusTooManyRequests,              // 429
	"InternalServerError":          http.StatusInternalServerError,          // 500
	"NotImplemented":               http.StatusNotImplemented,               // 501
	"BadGateway":                   http.StatusBadGateway,                   // 502
	"ServiceUnavailable":           http.StatusServiceUnavailable,           // 503
	"GatewayTimeout":               http.StatusGatewayTimeout,               // 504
	"HTTPVersionNotSupported":      http.StatusHTTPVersionNotSupported,      // 505
}

type Error struct {
	StatusCode int
	ErrorType  string
	Message    string
	Err        error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%v] %v: %v", e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("[%v] %v", e.StatusCode, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func New(errorType, message string) error {
	return &Error{StatusCode: statusMap[errorType], ErrorType: errorType, Message: message}
}

func Wrap(errorType, message string, err error) error {
	return &Error{StatusCode: statusMap[errorType], ErrorType: errorType, Message: message, Err: err}
}

func Is(err error, errorType string) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.StatusCode == statusMap[errorType]
	}
	return false
}
