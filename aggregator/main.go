package main

import (
	"context"
	"github.com/zoninnik89/ad-click-aggregator/aggregator/gateway"
	_ "github.com/zoninnik89/ad-click-aggregator/aggregator/gateway"
	common "github.com/zoninnik89/commons"
	"github.com/zoninnik89/commons/discovery"
	"github.com/zoninnik89/commons/discovery/consul"
	"google.golang.org/grpc"
	"net"
	"time"

	_ "github.com/joho/godotenv/autoload"
	zap "go.uber.org/zap"
	_ "google.golang.org/grpc"
)

var (
	serviceName = "aggregator"
	grpcAddr    = common.EnvString("GRPC_ADDR", "localhost:2000")
	consulAddr  = common.EnvString("CONSUL_ADDR", "localhost:8500")
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	zap.ReplaceGlobals(logger)

	registry, err := consul.NewRegistry(consulAddr, serviceName)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	instanceId := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceId, serviceName, grpcAddr); err != nil {
		panic(err)
	}

	go func() {
		for {
			if err := registry.HealthCheck(instanceId, serviceName); err != nil {
				logger.Error("Failed to health check", zap.Error(err))
			}
			time.Sleep(time.Second * 1)
		}
	}()

	defer registry.Deregister(ctx, instanceId, serviceName)

	grpcServer := grpc.NewServer()

	listner, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatal("failed to listen:", zap.Error(err))
	}
	defer listner.Close()

	gtw := gateway.NewGRPCGateway(registry)
	store := NewCountMinSketch(5, 20)
	service := NewService(store, gtw)

	NewGrpcHandler(grpcServer, service)

	logger.Info("Starting HTTP server", zap.String("port", grpcAddr))

	if err := grpcServer.Serve(listner); err != nil {
		logger.Fatal("failed to serve", zap.Error(err))
	}
}
