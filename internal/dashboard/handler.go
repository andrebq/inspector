package dashboard

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/andrebq/inspector/internal/manager"
)

type (
	rootHandler struct {
		mux *http.ServeMux
		api string

		events []*manager.IOEvent
		lock   sync.RWMutex
	}
)

var (
	tmpl = template.Must(template.New("__root__").Parse(pages))
)

func newRoot() *rootHandler {
	r := &rootHandler{
		api: "http://localhost:8082/",
	}
	r.mux = http.NewServeMux()
	r.mux.HandleFunc("/builtin/htmx.js", r.serveJS("htmx.js", htmxMin))
	r.mux.HandleFunc("/builtin/morphdom.js", r.serveJS("morphdom.js", morphdomMin))
	r.mux.HandleFunc("/index", r.index)
	r.mux.HandleFunc("/requests", r.requests)
	r.mux.HandleFunc("/inspect-request", r.inspectRequest)
	r.mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		r.index(w, req)
	})
	return r
}

func (r *rootHandler) renderTemplate(w http.ResponseWriter, req *http.Request, name string, tmplName string, data any) {
	buf := &bytes.Buffer{}
	err := tmpl.ExecuteTemplate(buf, tmplName, data)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Opsie!", http.StatusInternalServerError)
		return
	}
	// TODO: too lazy to properly implement a seek'able buffer
	http.ServeContent(w, req, name, time.Now(), strings.NewReader(buf.String()))
}

func (r *rootHandler) requests(w http.ResponseWriter, req *http.Request) {
	type item struct {
		Code int
		URL  string
		ID   int64
	}

	acc := []item{}
	r.lock.RLock()
	{
		for _, ev := range r.events {
			acc = append(acc, item{
				Code: ev.Code,
				URL:  ev.URL,
				ID:   ev.ID,
			})
		}
	}
	r.lock.RUnlock()

	r.renderTemplate(w, req, "requests.html", "requests", struct {
		Title    string
		Requests []item
	}{
		Title:    "Requests",
		Requests: acc,
	})
}

func (r *rootHandler) index(w http.ResponseWriter, req *http.Request) {
	r.renderTemplate(w, req, "index.html", "index", struct{ Title string }{Title: "Index"})
}

func (r *rootHandler) inspectRequest(w http.ResponseWriter, req *http.Request) {
	id, err := strconv.ParseInt(req.FormValue("rid"), 10, 64)
	if err != nil {
		http.Error(w, "invalid request id", http.StatusBadRequest)
		return
	}
	println(id)
}

func (r *rootHandler) serveJS(name, content string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		http.ServeContent(w, req, name, time.Now(), strings.NewReader(content))
	}
}

func (r *rootHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *rootHandler) fetchRequests(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		req, err := http.NewRequestWithContext(ctx, "GET", r.api, nil)
		if err != nil {
			return err
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Error establishing connection to inspector proxy api: %v", err)
			<-time.After(time.Minute)
			continue
		}
		if res.StatusCode != http.StatusOK {
			log.Printf("Unexpected response from server [%v - %v]", res.StatusCode, res.Status)
			<-time.After(time.Minute)
			continue
		}
		log.Printf("Connection with upstram server %v established", r.api)
		defer res.Body.Close()
		dec := json.NewDecoder(res.Body)
		for dec.More() {
			var out manager.IOEvent
			if err = dec.Decode(&out); errors.Is(err, io.EOF) {
				log.Printf("EOF from inspector proxy")
			} else if err != nil {
				log.Printf("Unexpected error form inspector proxy: %v", err)
			}
			r.lock.Lock()
			r.events = append(r.events, &out)
			r.lock.Unlock()
		}
		<-time.After(time.Minute)
	}
}
