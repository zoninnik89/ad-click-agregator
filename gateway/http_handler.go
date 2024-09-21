package main

import (
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/zoninnik89/ad-click-aggregator/gateway/logging"
	"github.com/zoninnik89/ad-click-aggregator/gateway/producer"
	store "github.com/zoninnik89/ad-click-aggregator/gateway/storage"
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
	cache   store.CacheInterface
}

func NewHandler(g gateway.AdGateway, q *producer.ClickProducer, c store.CacheInterface) *Handler {
	l := logging.GetLogger().Sugar()
	return &Handler{gateway: g, queue: q, logger: l, cache: c}
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

	h.cache.Put(adID) // store adID in cache to be used for GetCounter request to avoid excessive network calls

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

	h.cache.Put(ad.AdID) // store adID in cache to be used for GetCounter request to avoid excessive network calls

	h.logger.Infow("Ad was created, request returned", zap.Any("ad", ad))
	common.WriteJson(writer, http.StatusOK, ad)
}

func (h *Handler) HandleGetCounter(writer http.ResponseWriter, request *http.Request) {
	adID := request.URL.Query().Get("adID")
	advertiserID := request.URL.Query().Get("advertiserID")

	h.logger.Infow("GetCounter request received", "adID", adID)

	ctx := request.Context()

	// Check that ad exists, first in cache, then if it's not in local cache, check in DB in ads service
	if err := h.checkIfAdExists(adID, advertiserID); err != nil {
		h.logger.Errorw("Request returned error", zap.Error(err))
		common.WriteError(writer, http.StatusNotFound, err.Error())
		return
	} else {
		h.cache.Put(adID)
		h.logger.Infow("AdID was successfully retrieved and cache was updated", zap.Any("adID", adID))
	}

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
	log.Printf("SendClick request received")
	body, err := io.ReadAll(request.Body)
	if err != nil {
		h.logger.Errorw("Error reading request body", zap.Error(err))
		http.Error(writer, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer request.Body.Close()

	var requestData SendClickRequestData
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		h.logger.Errorw("Error unmarshalling request body", zap.Error(err))
		http.Error(writer, "Invalid JSON", http.StatusBadRequest)
		return
	}

	adID := requestData.AdID
	impressionID := requestData.ImpressionID
	ts := strconv.FormatInt(generateTS(), 10)

	h.logger.Infow("JSON from the request was successfully unmarshalled", "adID", adID, "impressionID", impressionID)

	value := adID + "," + impressionID + "," + ts

	deliveryChan := make(chan kafka.Event)
	err = h.queue.Publish(value, "clicks", nil, deliveryChan)

	if err != nil {
		h.logger.Errorw("Error publishing click event in Kafka", zap.Error(err))
		common.WriteJson(writer, http.StatusBadRequest, err)
	}

	e := <-deliveryChan
	msg := e.(*kafka.Message)

	if msg.TopicPartition.Error != nil {
		h.logger.Errorw("Message was not published", "error", msg.TopicPartition.Error)
	} else {
		h.logger.Infow("Message successfully published", "message", msg.TopicPartition, "time", msg.Timestamp.String())
	}
	close(deliveryChan)

	common.WriteJson(writer, http.StatusOK, ClickResponseData{AdID: adID, TS: ts})
}

func (h *Handler) checkIfAdExists(adID, advertiserID string) error {
	if adExistsInCache := h.cache.Get(adID); !adExistsInCache {
		if _, err := h.gateway.GetAd(ctx, advertiserID, adID); err != nil {
			return err
		}
	}
	return nil
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
