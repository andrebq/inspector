package manager

import "net/http"

type (
	// IOEvent represents either an incoming http request or
	// an outgoing http response
	IOEvent struct {
		ID      int64 `json:"id,omitempty"`
		Request struct {
			Body    string      `json:"body"`
			Headers http.Header `json:"headers"`
		} `json:"request,omitempty"`
		Response struct {
			Body    string      `json:"body"`
			Headers http.Header `json:"headers"`
		} `json:"Response,omitempty"`
		Code int    `json:"code,omitempty"`
		URL  string `json:"url,omitempty"`
	}
)
