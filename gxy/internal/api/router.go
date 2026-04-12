package api

import (
	"net/http"

	"github.com/ixpay-pro/gxy/internal/proxy"
)

type Router struct {
	handler *Handler
	proxy   *proxy.Proxy
}

func NewRouter(handler *Handler, proxy *proxy.Proxy) *Router {
	return &Router{
		handler: handler,
		proxy:   proxy,
	}
}

func (r *Router) SetupRoutes() {
	// API routes
	http.HandleFunc("/api//register", r.handler.RegisterService)
	http.HandleFunc("/api//deregister", r.handler.DeregisterService)
	http.HandleFunc("/api//services", r.handler.GetServices)

	// Proxy routes - all other routes are forwarded to services
	http.HandleFunc("/", r.proxy.ServeHTTP)
}
