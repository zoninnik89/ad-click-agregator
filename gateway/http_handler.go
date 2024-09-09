package main

import (
	"errors"
	"net/http"

	"github.com/zoninnik89/ad-click-aggregator/gateway/gateway"
	common "github.com/zoninnik89/commons"
	protoBuff "github.com/zoninnik89/commons/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	gateway gateway.AdGateway
}

func NewHandler(gateway gateway.AdGateway) *Handler {
	return &Handler{gateway}
}

func (handler *Handler) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/ads", handler.HandleCreateAd)
	mux.HandleFunc("GET /api/ads/{adID}", handler.handleGetAd)
}

func (handler *Handler) HandleGetAd(writer http.ResponseWriter, request *http.Request) {

}

func (handler *Handler) HandleCreateAd(writer http.ResponseWriter, request *http.Request) {
	adID := request.PathValue("adID")
	adTitle := request.PathValue("adTitle")

	if err := validateAd(adID, adTitle); err != nil {
		common.WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	ad, err := handler.client.CreateAd(request.Context(), &protoBuff.CreateAdRequest{
		AdId:    adID,
		AdTitle: adTitle,
	})

	rStatus := status.Convert(err)

	if rStatus != nil {
		if rStatus.Code() != codes.InvalidArgument {
			common.WriteJson(writer, http.StatusBadRequest, rStatus.Message())
			return
		}

		common.WriteError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	common.WriteJson(writer, http.StatusOK, ad)

}

func validateAd(adID string, adTitle string) error {
	if len(adID) == 0 {
		return errors.New("ad ID is missing")
	}
	if len(adTitle) == 0 {
		return errors.New("ad Title is missing")
	}
	return nil
}
