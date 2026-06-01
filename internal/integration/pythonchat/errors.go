package pythonchat

import "fmt"

type ErrorKind string

const (
	ErrorKindConfig         ErrorKind = "config"
	ErrorKindRequestBuild   ErrorKind = "request_build"
	ErrorKindCall           ErrorKind = "call"
	ErrorKindHTTPStatus     ErrorKind = "http_status"
	ErrorKindEnvelopeDecode ErrorKind = "envelope_decode"
	ErrorKindEnvelopeCode   ErrorKind = "envelope_code"
	ErrorKindEmptyData      ErrorKind = "empty_data"
)

type Error struct {
	Kind       ErrorKind
	StatusCode int
	Code       int
	Err        error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	switch {
	case e.StatusCode > 0:
		return fmt.Sprintf("python chat %s: http status %d", e.Kind, e.StatusCode)
	case e.Code != 0:
		return fmt.Sprintf("python chat %s: envelope code %d", e.Kind, e.Code)
	case e.Err != nil:
		return fmt.Sprintf("python chat %s: %v", e.Kind, e.Err)
	default:
		return fmt.Sprintf("python chat %s", e.Kind)
	}
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func newError(kind ErrorKind, err error) *Error {
	return &Error{Kind: kind, Err: err}
}
