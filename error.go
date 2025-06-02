package httputil

import (
	"net/http"
)

type StatusError struct {
	StatusCode int
}

func (s *StatusError) Error() string {
	return "error status: " + http.StatusText(s.StatusCode)
}
