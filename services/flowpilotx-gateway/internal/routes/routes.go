package routes

import (
	"github.com/flowpilotx/libs/config"
	"github.com/flowpilotx/libs/logger"
	"github.com/gorilla/mux"
)

// Router holds the mux router and dependencies
type Router struct {
	router *mux.Router
	log    logger.LoggerInterface
	cfg    *config.Config
}

// New creates a new router instance
func New(cfg *config.Config, log logger.LoggerInterface) *Router {
	return &Router{
		router: mux.NewRouter(),
		log:    log,
		cfg:    cfg,
	}
}

// Init initializes all routes
func (r *Router) Init() *mux.Router {
	// Initialize all route groups
	r.initAPIRoutes()
	r.initWebSocketRoutes()
	r.initMiddleware()

	return r.router
}

// GetRouter returns the underlying mux router
func (r *Router) GetRouter() *mux.Router {
	return r.router
}
