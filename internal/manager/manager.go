package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"sync"
	"sync/atomic"
)

type (
	requestData struct {
		ID       int64
		Req      *http.Request
		Response *httptest.ResponseRecorder
		Body     []byte
	}

	// M controls both the proxy redirection and the clients that want to inspect requests
	M struct {
		lock     sync.RWMutex
		probes   map[chan *requestData]struct{}
		Upstream *httputil.ReverseProxy

		rcount int64
	}
)

func (m *M) Proxy() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("X-Inspected", "true")
		requestID := m.inspectRequest(req)
		log := httptest.NewRecorder()
		m.Upstream.ServeHTTP(log, req)
		for k, vals := range log.Header() {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(log.Code)
		w.Write(log.Body.Bytes())
		go m.inspectResponse(requestID, log)
	})
}

func (m *M) Manager() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// disables browser 'smart content guessing'
		w.Header().Set("X-Content-Type-Options", "nosniff")
		flush, ok := w.(http.Flusher)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Response cannot be chunked!")
			return
		}
		probe := m.registerProbe()
		defer m.removeProbe(probe)
		for {
			select {
			case <-req.Context().Done():
				return
			case data, open := <-probe:
				if !open {
					return
				}

				var buf []byte
				signal := IOEvent{}
				if data.Response != nil {
					signal.ResponseID = data.ID
					signal.Code = data.Response.Code
					signal.Body = string(data.Body)
					signal.Headers = data.Response.Header()
				} else {
					signal.RequestID = data.ID
					signal.Headers = data.Req.Header
					signal.Body = string(data.Body)
				}
				buf, _ = json.Marshal(signal)
				_, err := w.Write(buf)
				if err != nil {
					return
				}
				fmt.Fprintln(w)
				flush.Flush()
			}
		}
	})
}

func (m *M) registerProbe() chan *requestData {
	m.lock.Lock()
	defer m.lock.Unlock()
	probe := make(chan *requestData, 1000)
	if m.probes == nil {
		m.probes = make(map[chan *requestData]struct{})
	}
	m.probes[probe] = struct{}{}
	return probe
}

func (m *M) removeProbe(p chan *requestData) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.probes == nil {
		return
	}
	delete(m.probes, p)
}

func (m *M) inspectResponse(rid int64, res *httptest.ResponseRecorder) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if len(m.probes) == 0 {
		return
	}
	rd := &requestData{
		ID:       rid,
		Response: res,
		Body:     res.Body.Bytes(),
	}

	for probe := range m.probes {
		// avoid blocking if probes are too slow to consume
		select {
		case probe <- rd:
		default:
		}
	}
}

func (m *M) inspectRequest(req *http.Request) int64 {
	rid := atomic.AddInt64(&m.rcount, 1)
	m.lock.RLock()
	defer m.lock.RUnlock()
	if len(m.probes) == 0 {
		return rid
	}
	copy := req.Clone(req.Context())
	body, _ := io.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewBuffer(body))

	rd := &requestData{
		ID:   rid,
		Req:  copy,
		Body: body,
	}

	for probe := range m.probes {
		// avoid blocking if probes are too slow to consume
		select {
		case probe <- rd:
		default:
		}
	}
	return rid
}
