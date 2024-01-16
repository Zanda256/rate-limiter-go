package rlgroup

import (
	"context"
	"net/http"

	"github.com/Zanda256/rate-limiter-go/foundation/logger"
	"github.com/Zanda256/rate-limiter-go/foundation/web"
)

// Handlers manages the set of check endpoints.
type Handlers struct {
	log *logger.Logger
}

// New constructs a handlers for route access.
func New(log *logger.Logger) *Handlers {
	return &Handlers{
		log: log,
	}
}

// type Handle func(http.ResponseWriter, *http.Request, httprouter.Params)
func (h *Handlers) Limited(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
	return web.Respond(context.Background(), rw, "Limited, don't over use me!", http.StatusOK)
}

func (h *Handlers) UnLimited(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
	h.log.Info(ctx, "Hit unlimited endpoint")
	return web.Respond(context.Background(), rw, "Unlimited! Let's Go!", http.StatusOK)
}
