package main

import (
	"context"
	store "github.com/zoninnik89/ad-click-aggregator/aggregator/storage"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/zoninnik89/ad-click-aggregator/gateway/gateway"
	"github.com/zoninnik89/ad-click-aggregator/gateway/logging"
	kafkaProducer "github.com/zoninnik89/ad-click-aggregator/gateway/producer"
	common "github.com/zoninnik89/commons"
	"github.com/zoninnik89/commons/discovery"
	"github.com/zoninnik89/commons/discovery/consul"
)

var (
	serviceName   = "gateway"
	httpAddress   = common.EnvString("HTTP_ADDR", ":8080")
	consulAddress = common.EnvString("CONSUL_ADDR", ":8500")
)

func main() {
	// add open telemetry

	logger := logging.InitLogger()
	defer logging.Sync()

	registry, err := consul.NewRegistry(consulAddress, serviceName)
	if err != nil {
		logger.Panic("Failed to connect to Consul", zap.Error(err))
		panic(err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, httpAddress); err != nil {
		logger.Panic("Failed to register service", zap.Error(err))
		panic(err)
	}

	go func() {
		for {
			if err := registry.HealthCheck(instanceID, serviceName); err != nil {
				logger.Warn("Failed to health check", zap.Error(err))
			}
			time.Sleep(time.Second * 1)
		}
	}()

	defer func(registry *consul.Registry, ctx context.Context, instanceID string, serviceName string) {
		err := registry.Deregister(ctx, instanceID, serviceName)
		if err != nil {
			logger.Fatal("Failed to deregister service", zap.Error(err))
		}
	}(registry, ctx, instanceID, serviceName)

	mux := http.NewServeMux()
	adsGateway := gateway.NewGRPCGateway(registry)

	logger.Info("Starting Kafka Producer")

	kp := kafkaProducer.NewKafkaProducer()
	defer kp.Producer.Flush(10)

	cacheCap, _ := strconv.Atoi(common.EnvString("CACHE_CAP", "500"))
	cache := store.NewCache(cacheCap)

	handler := NewHandler(adsGateway, kp, cache)
	handler.registerRoutes(mux)

	logger.Info("Starting HTTP server at %s", zap.String("port", httpAddress))

	if err := http.ListenAndServe(httpAddress, mux); err != nil {
		logger.Fatal("Failed to start http server")
	}

}
