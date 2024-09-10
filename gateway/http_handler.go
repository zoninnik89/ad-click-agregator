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
	mux.HandleFunc("GET /api/ads/{adID}", handler.HandleGetAd)
}

func (handler *Handler) HandleGetAd(writer http.ResponseWriter, request *http.Request) {
	advertiserID := request.PathValue("advertiserID")
	adID := request.PathValue("adID")

	ctx := request.Context()
	ad, err := handler.gateway.GetAd(ctx, advertiserID, adID)
	rStatus := status.Convert(err)
	if rStatus != nil {
		if rStatus.Code() != codes.InvalidArgument {
			common.WriteError(writer, http.StatusBadRequest, rStatus.Message())
			return
		}

		common.WriteError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	common.WriteJson(writer, http.StatusOK, ad)
}

func (handler *Handler) HandleCreateAd(writer http.ResponseWriter, request *http.Request) {
	adID := request.PathValue("ID")
	advertiserID := request.PathValue("AdvertiserID")
	adTitle := request.PathValue("Title")
	adURL := request.PathValue("AdURL")

	if err := validateAd(adID, adTitle); err != nil {
		common.WriteError(writer, http.StatusBadRequest, err.Error())
		return
	}

	ad, err := handler.gateway.CreateAd(request.Context(), &protoBuff.CreateAdRequest{
		ID:           adID,
		AdvertiserID: advertiserID,
		Title:        adTitle,
		AdURL:        adURL,
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
