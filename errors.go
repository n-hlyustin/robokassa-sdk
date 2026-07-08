package robokassa

import "fmt"

type SDKError struct {
	Op         string
	StatusCode int
	Message    string
	Err        error
}

func (e *SDKError) Error() string {
	switch {
	case e == nil:
		return ""
	case e.StatusCode > 0 && e.Err != nil:
		return fmt.Sprintf("%s: %s (status=%d): %v", e.Op, e.Message, e.StatusCode, e.Err)
	case e.StatusCode > 0:
		return fmt.Sprintf("%s: %s (status=%d)", e.Op, e.Message, e.StatusCode)
	case e.Err != nil:
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
	default:
		return fmt.Sprintf("%s: %s", e.Op, e.Message)
	}
}

func (e *SDKError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
