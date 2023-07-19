package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"
)

var (
	upstream  = flag.String("upstream", "http://localhost:8081", "Upstream URL")
	proxyAddr = flag.String("proxy-addr", "127.0.0.1:8080", "Address to bind and listen for incoming HTTP requests")
	mngAddr   = flag.String("mng-addr", "127.0.0.1:8082", "Address to bind and list for proxy management commands")
)

type (
	requestData struct {
		ID       int64
		Req      *http.Request
		Response *httptest.ResponseRecorder
		Body     []byte
	}

	manager struct {
		lock     sync.RWMutex
		probes   map[chan *requestData]struct{}
		upstream *httputil.ReverseProxy

		rcount int64
	}
)

func (m *manager) proxy(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("X-Inspected", "true")
	requestID := m.inspectRequest(req)
	log := httptest.NewRecorder()
	m.upstream.ServeHTTP(log, req)
	for k, vals := range log.Header() {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(log.Code)
	w.Write(log.Body.Bytes())
	go m.inspectResponse(requestID, log)
}

func (m *manager) manager(w http.ResponseWriter, req *http.Request) {
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
			signal := struct {
				RequestID  int64       `json:"requestId,omitempty"`
				ResponseID int64       `json:"responseId,omitempty"`
				Body       string      `json:"body"`
				Headers    http.Header `json:"headers"`
				Code       int         `json:"code,omitempty"`
			}{}
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
}

func (m *manager) registerProbe() chan *requestData {
	m.lock.Lock()
	defer m.lock.Unlock()
	probe := make(chan *requestData, 1000)
	if m.probes == nil {
		m.probes = make(map[chan *requestData]struct{})
	}
	m.probes[probe] = struct{}{}
	return probe
}

func (m *manager) removeProbe(p chan *requestData) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.probes == nil {
		return
	}
	delete(m.probes, p)
}

func (m *manager) inspectResponse(rid int64, res *httptest.ResponseRecorder) {
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

func (m *manager) inspectRequest(req *http.Request) int64 {
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

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	flag.Parse()
	remote, err := url.Parse(*upstream)
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	mng := &manager{
		upstream: proxy,
	}

	proxyServer := &http.Server{
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
		Handler: http.HandlerFunc(mng.proxy),
		Addr:    *proxyAddr,
	}
	mngServer := &http.Server{
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
		Handler: http.HandlerFunc(mng.manager),
		Addr:    *mngAddr,
	}

	go serve(proxyServer, cancel)
	go serve(mngServer, cancel)
	<-ctx.Done()

	gracefulShutdown(time.Minute, proxyServer, mngServer)
}

func gracefulShutdown(timeout time.Duration, servers ...*http.Server) {
	if len(servers) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	wg := &sync.WaitGroup{}
	wg.Add(len(servers))
	for _, s := range servers {
		go func(s *http.Server) {
			s.Shutdown(ctx)
		}(s)
	}
	wg.Wait()
}

func serve(srv *http.Server, done func()) {
	defer done()
	err := srv.ListenAndServe()
	if err != nil {
		log.Printf("Server %v: %v", srv.Addr, err)
	}
}
