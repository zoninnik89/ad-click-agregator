package main

import (
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/zoninnik89/ad-click-aggregator/gateway/producer"
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

type ClickResponseData struct {
	AdID string `json:"adID"`
	TS   string `json:"ts"`
}

func NewHandler(gateway gateway.AdGateway, queue *producer.ClickProducer) *Handler {
	return &Handler{gateway: gateway, queue: queue}
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
	ts := strconv.FormatInt(generateTS(), 10)

	log.Printf("click request has adID: %v and impressionID: %v", adID, impressionID)

	value := adID + "," + impressionID + "," + ts

	deliveryChan := make(chan kafka.Event)
	err = handler.queue.Publish(value, "clicks", nil, deliveryChan)

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
