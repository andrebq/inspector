package proxy

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/andrebq/inspector/internal/dashboard"
	"github.com/andrebq/inspector/internal/manager"
	"github.com/urfave/cli/v3"
)

func gracefulShutdown(timeout time.Duration, servers ...*http.Server) error {
	if len(servers) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	wg := &sync.WaitGroup{}
	wg.Add(len(servers))
	shutdown := func(ctx context.Context, wg *sync.WaitGroup) func(*http.Server) {
		return func(s *http.Server) {
			defer wg.Done()
			s.Shutdown(ctx)
		}
	}
	for _, s := range servers {
		go shutdown(ctx, wg)(s)
	}
	wg.Wait()
	return ctx.Err()
}

func serve(srv *http.Server, done func()) {
	defer done()
	err := srv.ListenAndServe()
	if err != nil {
		log.Printf("Server %v: %v", srv.Addr, err)
	}
}

func Cmd() *cli.Command {
	var upstream string
	proxyAddr, mngAddr := "localhost:8081", "localhost:8082"
	dashboardAddr := "off"

	return &cli.Command{
		Name:  "proxy",
		Usage: "Runs the reverse proxy",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "upstream",
				Aliases:     []string{"u"},
				Usage:       "Upstream address of the HTTP request",
				DefaultText: "Eg.: http://localhost:8888",
				Required:    true,
				Destination: &upstream,
			},
			&cli.StringFlag{
				Name:        "proxy-addr",
				Aliases:     []string{"p"},
				Usage:       "Address where the proxy itself will listen for connections",
				Value:       proxyAddr,
				Destination: &proxyAddr,
			},
			&cli.StringFlag{
				Name:        "management-addr",
				Aliases:     []string{"m", "mng-addr"},
				Usage:       "Address where the proxy management server will list for connections",
				Value:       mngAddr,
				Destination: &mngAddr,
			},
			&cli.StringFlag{
				Name:        "dashboard",
				Aliases:     []string{"d"},
				Usage:       "Sets the address of the dashboard, options are: (off|on|<ip>:<port>). Setting to on will use localhost:8083 as the bind address",
				Value:       dashboardAddr,
				Destination: &dashboardAddr,
			},
		},
		Action: func(appCtx *cli.Context) error {
			ctx, cancel := context.WithCancel(appCtx.Context)

			remote, err := url.Parse(upstream)
			if err != nil {
				panic(err)
			}

			proxy := httputil.NewSingleHostReverseProxy(remote)
			mng := &manager.M{
				Upstream: proxy,
			}

			proxyServer := &http.Server{
				BaseContext: func(l net.Listener) context.Context {
					return ctx
				},
				Handler: mng.Proxy(),
				Addr:    proxyAddr,
			}
			mngServer := &http.Server{
				BaseContext: func(l net.Listener) context.Context {
					return ctx
				},
				Handler: mng.Manager(),
				Addr:    mngAddr,
			}

			var servers []*http.Server
			go serve(proxyServer, cancel)
			go serve(mngServer, cancel)

			servers = append(servers, proxyServer, mngServer)

			dashboardAddr = strings.ToLower(dashboardAddr)
			switch dashboardAddr {
			case "on":
				dashboardAddr = "localhost:8083"
			case "off":
				dashboardAddr = ""
			}
			if dashboardAddr != "" {
				handler := dashboard.Handler(ctx)
				dsSrv := &http.Server{
					Handler:     handler,
					BaseContext: func(l net.Listener) context.Context { return ctx },
					Addr:        dashboardAddr,
				}
				go serve(dsSrv, cancel)
				servers = append(servers, dsSrv)
			}
			<-ctx.Done()

			return gracefulShutdown(time.Minute, servers...)
		},
	}
}
