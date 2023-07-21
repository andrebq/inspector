package dashboard

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"
)

func Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	handler := newRoot()
	go func() {
		defer cancel()
		if err := handler.fetchRequests(ctx); err != nil {
			log.Printf("Error fetching requests: %v", err)
		}
	}()
	srv := &http.Server{
		Addr:    "localhost:8083",
		Handler: handler,
	}
	go func() {
		defer cancel()
		log.Printf("Starting dahsboard at %v", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Print(err)
		}
	}()
	<-ctx.Done()
	return gracefulShutdown(srv)
}

func gracefulShutdown(srv *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	return srv.Shutdown(ctx)
}
