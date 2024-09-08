package main

import (
	"net/http"

	protoBuff "github.com/zoninnik89/commons/api"
)

type Handler struct {
	client protoBuff.AdsServiceClient
}

func NewHandler(client protoBuff.AdsServiceClient) *Handler {
	return &Handler{client}
}

func (handler Handler) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/ads", handler.HandleCreateAd)
}

func (handler Handler) HandleCreateAd(writer http.ResponseWriter, request *http.Request) {
	adID := request.PathValue("adID")
	adTitle := request.PathValue("adtitle")

	handler.client.CreateAd(request.Context(), &protoBuff.CreateAdRequest{
		AdId:    adID,
		AdTitle: adTitle,
	})

}
