package manager

import "net/http"

type (
	// IOEvent represents either an incoming http request or
	// an outgoing http response
	IOEvent struct {
		RequestID  int64       `json:"requestId,omitempty"`
		ResponseID int64       `json:"responseId,omitempty"`
		Body       string      `json:"body"`
		Headers    http.Header `json:"headers"`
		Code       int         `json:"code,omitempty"`
	}
)
