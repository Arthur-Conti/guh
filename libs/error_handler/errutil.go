package errorhandler

import (
	"errors"
	"fmt"
	"net/http"
)

type ErrorType string

var (
	BadRequest                   ErrorType = "BadRequest"
	Unauthorized                 ErrorType = "Unauthorized"
	PaymentRequired              ErrorType = "PaymentRequired"
	Forbidden                    ErrorType = "Forbidden"
	NotFound                     ErrorType = "NotFound"
	MethodNotAllowed             ErrorType = "MethodNotAllowed"
	NotAcceptable                ErrorType = "NotAcceptable"
	ProxyAuthRequired            ErrorType = "Proxy AuthRequired"
	RequestTimeout               ErrorType = "RequestTimeout"
	Conflict                     ErrorType = "Conflict"
	Gone                         ErrorType = "Gone"
	LengthRequired               ErrorType = "LengthRequired"
	PreconditionFailed           ErrorType = "PreconditionFailed"
	RequestEntityTooLarge        ErrorType = "RequestEntityTooLarge"
	RequestURITooLong            ErrorType = "RequestURITooLong"
	UnsupportedMediaType         ErrorType = "UnsupportedMediaType"
	RequestedRangeNotSatisfiable ErrorType = "RequestedRangeNotSatisfiable"
	ExpectationFailed            ErrorType = "ExpectationFailed"
	Teapot                       ErrorType = "Teapot"
	UnprocessableEntity          ErrorType = "UnprocessableEntity"
	TooManyRequests              ErrorType = "TooManyRequests"
	InternalServerError          ErrorType = "InternalServerError"
	NotImplemented               ErrorType = "NotImplemented"
	BadGateway                   ErrorType = "BadGateway"
	ServiceUnavailable           ErrorType = "ServiceUnavailable"
	GatewayTimeout               ErrorType = "GatewayTimeout"
	HTTPVersionNotSupported      ErrorType = "HTTPVersionNotSupported"
)

var statusMap = map[ErrorType]int{
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
	ErrorType  ErrorType
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

func New(errorType ErrorType, message string) error {
	return &Error{StatusCode: statusMap[errorType], ErrorType: errorType, Message: message}
}

func Wrap(errorType ErrorType, message string, err error) error {
	return &Error{StatusCode: statusMap[errorType], ErrorType: errorType, Message: message, Err: err}
}

func Is(err error, errorType ErrorType) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.StatusCode == statusMap[errorType]
	}
	return false
}
