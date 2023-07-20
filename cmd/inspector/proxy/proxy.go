package proxy

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

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
	for _, s := range servers {
		go func(s *http.Server) {
			s.Shutdown(ctx)
		}(s)
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

			go serve(proxyServer, cancel)
			go serve(mngServer, cancel)
			<-ctx.Done()

			return gracefulShutdown(time.Minute, proxyServer, mngServer)
		},
	}
}
