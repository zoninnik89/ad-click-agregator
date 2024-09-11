package main

import (
	"context"
	"fmt"
	"github.com/zoninnik89/ad-click-aggregator/ads/gateway"
	"github.com/zoninnik89/commons/discovery"
	"github.com/zoninnik89/commons/discovery/consul"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/grpc"
	"net"
	"time"

	_ "github.com/zoninnik89/ad-click-aggregator/ads/gateway"
	common "github.com/zoninnik89/commons"
	_ "github.com/zoninnik89/commons/discovery"
	_ "github.com/zoninnik89/commons/discovery/consul"

	_ "github.com/joho/godotenv/autoload"
	_ "go.mongodb.org/mongo-driver/mongo"
	_ "go.mongodb.org/mongo-driver/mongo/options"
	_ "go.mongodb.org/mongo-driver/mongo/readpref"
	zap "go.uber.org/zap"
	_ "google.golang.org/grpc"
)

var (
	serviceName = "ads"
	grpcAddr    = common.EnvString("GRPC_ADDR", "localhost:2000")
	consulAddr  = common.EnvString("CONSUL_ADDR", "localhost:8500")
	mongoUser   = common.EnvString("MONGO_DB_USER", "root")
	mongoPass   = common.EnvString("MONGO_DB_PASS", "example")
	mongoAddr   = common.EnvString("MONGO_DB_HOST", "localhost:27017")
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

	// mongo connection
	uri := fmt.Sprintf("mongodb://%s:%s@%s", mongoUser, mongoPass, mongoAddr)
	mongoClient, err := connectToMongoDB(uri)
	if err != nil {
		logger.Fatal("failed to connect to mongodb", zap.Error(err))
	}

	grpcServer := grpc.NewServer()

	listner, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatalf("failed to listen:", zap.Error(err))
	}
	defer listner.Close()

	gtw := gateway.NewGRPCGateway(registry)

	store := NewStore(mongoClient)
	service := NewService(store, gtw)

	NewGrpcHandler(grpcServer, service)

	logger.Info("Starting HTTP server", zap.String("port", grpcAddr))

	if err := grpcServer.Serve(listner); err != nil {
		logger.Fatal("failed to serve", zap.Error(err))
	}

}

func connectToMongoDB(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, readpref.Primary())
	return client, err
}
