package rl_group

import (
	"github.com/Zanda256/rate-limiter-go/foundation/logger"
	"github.com/julienschmidt/httprouter"
	"net/http"
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
func (h *Handlers) Limited(http.ResponseWriter, *http.Request, httprouter.Params) {

}

func (h *Handlers) UnLimited(http.ResponseWriter, *http.Request, httprouter.Params) {

}
