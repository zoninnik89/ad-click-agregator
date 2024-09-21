package main

import (
	"context"
	c "github.com/zoninnik89/ad-click-aggregator/aggregator/consumer"
	"github.com/zoninnik89/ad-click-aggregator/aggregator/logging"
	"go.uber.org/zap"

	store "github.com/zoninnik89/ad-click-aggregator/aggregator/storage"
	common "github.com/zoninnik89/commons"
	"github.com/zoninnik89/commons/discovery"
	"github.com/zoninnik89/commons/discovery/consul"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	_ "google.golang.org/grpc"
)

var (
	serviceName = "aggregator"
	grpcAddr    = common.EnvString("GRPC_ADDR", "localhost:2001")
	consulAddr  = common.EnvString("CONSUL_ADDR", "localhost:8500")
)

func main() {
	logger := logging.InitLogger()
	defer logging.Sync()

	// Connect to Consul
	registry, err := consul.NewRegistry(consulAddr, serviceName)
	if err != nil {
		logger.Panic("Failed to connect to Consul", zap.Error(err))
		panic(err)
	}

	// Register self in Consul
	ctx := context.Background()
	instanceId := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceId, serviceName, grpcAddr); err != nil {
		logger.Panic("Failed to register service", zap.Error(err))
		panic(err)
	}

	// Run the service and send health checks
	go func() {
		for {
			if err := registry.HealthCheck(instanceId, serviceName); err != nil {
				logger.Warn("Failed to health check", zap.Error(err))
			}
			time.Sleep(time.Second * 1)
		}
	}()

	// Deferring the deregister call until the service is shut down
	defer func(registry *consul.Registry, ctx context.Context, instanceID string, serviceName string) {
		err := registry.Deregister(ctx, instanceID, serviceName)
		if err != nil {
			logger.Fatal("Failed to deregister service", zap.Error(err))
		}
	}(registry, ctx, instanceId, serviceName)

	// Initialize the grpc server
	grpcServer := grpc.NewServer()

	l, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatal("Failed to listen:", zap.Error(err))
	}
	defer func(l net.Listener) {
		err := l.Close()
		if err != nil {
			logger.Warn("Failed to close listener", zap.Error(err))
		}
	}(l)

	storage := store.NewCountMinSketch(5, 20)
	cacheCap, _ := strconv.Atoi(common.EnvString("CACHE_CAP", "500"))
	cache := store.NewCache(cacheCap)
	service := NewService(storage, cache)

	NewGrpcHandler(grpcServer, service)

	logger.Info("Starting HTTP server", zap.String("port", grpcAddr))

	logger.Info("Starting Kafka Consumer")
	consumer := c.NewKafkaConsumer()

	topics := []string{"clicks"}
	err = consumer.SubscribeTopics(topics, nil)

	if err != nil {
		panic(err)
	}

	go func() {
		click, err := service.ConsumeClick(ctx, consumer)
		if err != nil {
			logger.Fatal("Failed to consume click: %v", zap.Error(err))
		} else {
			logger.Info("click was added to the store", zap.String("ad", click.AdID), zap.String("timestamp", click.Timestamp))
		}

		time.Sleep(time.Second * 1)
	}()

	if err := grpcServer.Serve(l); err != nil {
		logger.Fatal("Failed to serve", zap.Error(err))
	}
}
