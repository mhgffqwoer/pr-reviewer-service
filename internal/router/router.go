package router

import (
	"net/http"
)

type Router struct {
	mux *http.ServeMux
}

func New() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt.mux.ServeHTTP(w, r)
}

func (rt *Router) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	rt.mux.HandleFunc(pattern, handler)
}
