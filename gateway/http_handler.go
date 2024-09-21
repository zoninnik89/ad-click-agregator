package main

import (
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/zoninnik89/ad-click-aggregator/gateway/logging"
	"github.com/zoninnik89/ad-click-aggregator/gateway/producer"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/zoninnik89/ad-click-aggregator/gateway/gateway"
	common "github.com/zoninnik89/commons"
	protoBuff "github.com/zoninnik89/commons/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	gateway gateway.AdGateway
	queue   *producer.ClickProducer
	logger  *zap.SugaredLogger
}

func NewHandler(gateway gateway.AdGateway, queue *producer.ClickProducer) *Handler {
	l := logging.GetLogger().Sugar()
	return &Handler{gateway: gateway, queue: queue, logger: l}
}

func (h *Handler) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/ads", h.HandleCreateAd)
	mux.HandleFunc("GET /api/ads", h.HandleGetAd)
	mux.HandleFunc("GET /api/counter/", h.HandleGetCounter)
	mux.HandleFunc("POST /api/sendclick", h.HandleSendClick)
}

func (h *Handler) HandleGetAd(writer http.ResponseWriter, request *http.Request) {
	advertiserID := request.URL.Query().Get("advertiserID")
	adID := request.URL.Query().Get("adID")

	h.logger.Infow("GetAd request received", "adID", adID, "advertiserID", advertiserID)

	ctx := request.Context()
	ad, err := h.gateway.GetAd(ctx, advertiserID, adID)
	rStatus := status.Convert(err)
	if rStatus != nil {
		if rStatus.Code() != codes.InvalidArgument {
			h.logger.Errorw("Request returned error", zap.String("message", rStatus.Message()))
			common.WriteError(writer, http.StatusBadRequest, rStatus.Message())
			return
		}
		h.logger.Errorw("Request returned error", zap.Error(err))
		common.WriteError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.Infow("Request returned", zap.Any("ad", ad))
	common.WriteJson(writer, http.StatusOK, ad)
}

func (h *Handler) HandleCreateAd(writer http.ResponseWriter, request *http.Request) {
	h.logger.Infow("CreateAd request received")

	body, err := io.ReadAll(request.Body)
	if err != nil {
		h.logger.Errorw("Error reading request body", zap.Error(err))
		http.Error(writer, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer request.Body.Close()

	var requestData CreateAdRequestData
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		h.logger.Errorw("Error unmarshalling request body", zap.Error(err))
		http.Error(writer, "Invalid JSON", http.StatusBadRequest)
		return
	}

	advertiserID := requestData.AdvertiserID
	adTitle := requestData.AdTitle
	adURL := requestData.AdURL

	h.logger.Infow("Request body successfully unmarshalled", "advertiserID", advertiserID, "adTitle", adTitle, "adURL", adURL)

	//if err := validateAd(adID, adTitle); err != nil {
	//	common.WriteError(writer, http.StatusBadRequest, err.Error())
	//	return
	//}

	ad, err := h.gateway.CreateAd(request.Context(), &protoBuff.CreateAdRequest{
		AdvertiserID: advertiserID,
		Title:        adTitle,
		AdURL:        adURL,
	})

	rStatus := status.Convert(err)

	if rStatus != nil {
		if rStatus.Code() != codes.InvalidArgument {
			h.logger.Errorw("Request returned error", zap.Error(err))
			common.WriteJson(writer, http.StatusBadRequest, rStatus.Message())
			return
		}

		h.logger.Errorw("Request returned error", zap.Error(err))
		common.WriteError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.Infow("Ad was created, request returned", zap.Any("ad", ad))
	common.WriteJson(writer, http.StatusOK, ad)

}

func (h *Handler) HandleGetCounter(writer http.ResponseWriter, request *http.Request) {
	adID := request.URL.Query().Get("adID")
	h.logger.Infow("GetCounter request received", "adID", adID)

	ctx := request.Context()

	// Check that ad exists, first in cache, then if it's not in local cache, check in DB in ads service

	ad, err := h.gateway.GetClickCounter(ctx, adID)
	rStatus := status.Convert(err)
	if rStatus != nil {
		if rStatus.Code() != codes.InvalidArgument {
			h.logger.Errorw("Request returned error", zap.Error(err))
			common.WriteError(writer, http.StatusBadRequest, rStatus.Message())
			return
		}

		h.logger.Errorw("Request returned error", zap.Error(err))
		common.WriteError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.Infow("Ad exists, counter request returned", zap.Any("ad", ad))
	common.WriteJson(writer, http.StatusOK, ad)
}

func (h *Handler) HandleSendClick(writer http.ResponseWriter, request *http.Request) {
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
	ts := strconv.FormatInt(generateTS(), 10)

	log.Printf("click request has adID: %v and impressionID: %v", adID, impressionID)

	value := adID + "," + impressionID + "," + ts

	deliveryChan := make(chan kafka.Event)
	err = h.queue.Publish(value, "clicks", nil, deliveryChan)

	if err != nil {
		common.WriteJson(writer, http.StatusBadRequest, err)
	}

	e := <-deliveryChan
	msg := e.(*kafka.Message)

	if msg.TopicPartition.Error != nil {
		log.Printf("Message was not published, error: %v", msg.TopicPartition.Error)
	} else {
		log.Printf("Message published in: %v", msg.TopicPartition)
	}

	close(deliveryChan)
	log.Printf("click request with adID: %v, was stored at: %v ", adID, ts)

	common.WriteJson(writer, http.StatusOK, ClickResponseData{AdID: adID, TS: ts})
}

func generateTS() int64 {
	currentTime := time.Now().Unix()
	return currentTime
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
