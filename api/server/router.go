package server

import "net/http"

type routeInfo struct {
	route   string
	method  string
	headers *[]string
}

type Router struct {
	routes map[routeInfo]http.HandlerFunc
}

func (r *Router) HandleFunc(route, method string, headers []string, h http.HandlerFunc) {
	r.routes[routeInfo{route, method, &headers}] = h
}
