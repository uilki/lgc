package server

import "net/http"

type Router struct {
	routes map[string]http.HandlerFunc
}

func (r *Router) HandleFunc(route string, h http.HandlerFunc) {
	r.routes[route] = h
}
