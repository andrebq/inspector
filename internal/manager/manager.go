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
		probes   map[chan *IOEvent]struct{}
		Upstream *httputil.ReverseProxy

		rcount int64
	}
)

func (m *M) Proxy() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !m.hasProbes() {
			// bypass all the processing since nobody is looking at the data
			m.Upstream.ServeHTTP(w, req)
			return
		}
		w.Header().Set("X-Inspected", "true")
		ev := m.inspectRequest(req)
		log := httptest.NewRecorder()
		m.Upstream.ServeHTTP(log, req)
		for k, vals := range log.Header() {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(log.Code)
		w.Write(log.Body.Bytes())
		go m.inspectResponse(ev, log)
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
			case ev, open := <-probe:
				if !open {
					return
				}
				buf, _ := json.Marshal(ev)
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

func (m *M) registerProbe() chan *IOEvent {
	m.lock.Lock()
	defer m.lock.Unlock()
	probe := make(chan *IOEvent, 1000)
	if m.probes == nil {
		m.probes = make(map[chan *IOEvent]struct{})
	}
	m.probes[probe] = struct{}{}
	return probe
}

func (m *M) removeProbe(p chan *IOEvent) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.probes == nil {
		return
	}
	delete(m.probes, p)
}

func (m *M) inspectResponse(ev *IOEvent, res *httptest.ResponseRecorder) {
	if ev == nil {
		return
	}
	m.lock.RLock()
	defer m.lock.RUnlock()
	if len(m.probes) == 0 {
		return
	}
	ev.Code = res.Code
	// TODO: change this to use bytes instead
	ev.Response.Body = res.Body.String()
	ev.Response.Headers = res.Header()

	for probe := range m.probes {
		// avoid blocking if probes are too slow to consume
		select {
		case probe <- ev:
		default:
		}
	}
}

func (m *M) hasProbes() bool {
	m.lock.RLock()
	val := len(m.probes) > 0
	m.lock.RUnlock()
	return val
}

func (m *M) inspectRequest(req *http.Request) *IOEvent {
	rid := atomic.AddInt64(&m.rcount, 1)
	body, _ := io.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewBuffer(body))

	ev := &IOEvent{
		ID:   rid,
		Code: 0,
		URL:  req.URL.String(),
	}
	ev.Request.Body = string(body)
	ev.Request.Headers = req.Header
	return ev
}
