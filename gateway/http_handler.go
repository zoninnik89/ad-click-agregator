package main

import "net/http"

type Handler struct {
	// gateway
}

func NewHandler() *Handler {
	return &Handler{}
}

func (handler Handler) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/ads", handler.HandleCreateAd)
}

func (handler Handler) HandleCreateAd(writer http.ResponseWriter, request *http.Request) {

}
