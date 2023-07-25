package dashboard

import (
	"context"
	"log"
	"net/http"
)

func Handler(ctx context.Context) http.Handler {
	handler := newRoot()
	go func() {
		if err := handler.fetchRequests(ctx); err != nil {
			log.Printf("Error fetching requests: %v", err)
		}
	}()
	return handler
}
