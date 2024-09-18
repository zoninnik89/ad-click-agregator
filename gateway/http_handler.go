package main

import (
	"encoding/json"
	"io"
	"log"
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

type CreateAdRequestData struct {
	AdvertiserID string `json:"advertiserID"`
	AdTitle      string `json:"adTitle"`
	AdURL        string `json:"adURL"`
}

type SendClickRequestData struct {
	AdID         string `json:"adID"`
	ImpressionID string `json:"impressionID"`
}

func NewHandler(gateway gateway.AdGateway) *Handler {
	return &Handler{gateway}
}

func (handler *Handler) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/ads", handler.HandleCreateAd)
	mux.HandleFunc("GET /api/ads", handler.HandleGetAd)
	mux.HandleFunc("GET /api/counter/", handler.HandleGetCounter)
	mux.HandleFunc("POST /api/sendclick", handler.HandleSendClick)
}

func (handler *Handler) HandleGetAd(writer http.ResponseWriter, request *http.Request) {
	advertiserID := request.URL.Query().Get("advertiserID")
	adID := request.URL.Query().Get("adID")

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
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer request.Body.Close()

	var requestData CreateAdRequestData
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		http.Error(writer, "Invalid JSON", http.StatusBadRequest)
		return
	}

	advertiserID := requestData.AdvertiserID
	adTitle := requestData.AdTitle
	adURL := requestData.AdURL

	//if err := validateAd(adID, adTitle); err != nil {
	//	common.WriteError(writer, http.StatusBadRequest, err.Error())
	//	return
	//}

	ad, err := handler.gateway.CreateAd(request.Context(), &protoBuff.CreateAdRequest{
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

func (handler *Handler) HandleGetCounter(writer http.ResponseWriter, request *http.Request) {
	adID := request.URL.Query().Get("adID")

	ctx := request.Context()
	log.Printf("Cur ad is: %v", adID)
	ad, err := handler.gateway.GetClickCounter(ctx, adID)
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

func (handler *Handler) HandleSendClick(writer http.ResponseWriter, request *http.Request) {
	log.Printf("received send click request")
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer request.Body.Close()

	var requestData SendClickRequestData
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		http.Error(writer, "Invalid JSON", http.StatusBadRequest)
		return
	}

	adID := requestData.AdID
	impressionID := requestData.ImpressionID

	log.Printf("click request has adID: %v and impressionID: %v", adID, impressionID)

	click, err := handler.gateway.SendClick(request.Context(), &protoBuff.SendClickRequest{
		AdID:         adID,
		ImpressionID: impressionID,
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

	log.Printf("click request with adID: %v, was stored at: %v ", click.AdID, click.Timestamp)

	common.WriteJson(writer, http.StatusOK, click)
}

//func validateAd(adID string, adTitle string) error {
//	if len(adID) == 0 {
//		return errors.New("ad ID is missing")
//	}
//	if len(adTitle) == 0 {
//		return errors.New("ad Title is missing")
//	}
//	return nil
//}
