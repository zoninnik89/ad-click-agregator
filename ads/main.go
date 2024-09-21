package main

import (
	"context"
	"fmt"
	"github.com/zoninnik89/ad-click-aggregator/ads/logging"
	"github.com/zoninnik89/commons/discovery"
	"github.com/zoninnik89/commons/discovery/consul"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"time"

	common "github.com/zoninnik89/commons"
	_ "github.com/zoninnik89/commons/discovery"
	_ "github.com/zoninnik89/commons/discovery/consul"

	_ "github.com/joho/godotenv/autoload"
	_ "go.mongodb.org/mongo-driver/mongo"
	_ "go.mongodb.org/mongo-driver/mongo/options"
	_ "go.mongodb.org/mongo-driver/mongo/readpref"
	_ "google.golang.org/grpc"
	_ "strconv"
)

var (
	serviceName = "ads"
	grpcAddr    = common.EnvString("GRPC_ADDR", "localhost:2000")
	consulAddr  = common.EnvString("CONSUL_ADDR", "localhost:8500")
	mongoUser   = common.EnvString("MONGO_DB_USER", "root")
	mongoPass   = common.EnvString("MONGO_DB_PASS", "rootpassword")
	mongoAddr   = common.EnvString("MONGO_DB_HOST", "localhost:27017")
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

	// MongoDB connection
	uri := fmt.Sprintf("mongodb://%s:%s@%s", mongoUser, mongoPass, mongoAddr)
	mongoClient, err := connectToMongoDB(uri)
	if err != nil {
		logger.Fatal("Failed to connect to mongodb", zap.Error(err))
	}

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

	// initialize the DB client and new service
	store := NewStore(mongoClient)
	service := NewService(store)

	NewGrpcHandler(grpcServer, service)

	logger.Info("Starting HTTP server", zap.String("port", grpcAddr))

	if err := grpcServer.Serve(l); err != nil {
		logger.Fatal("Failed to serve", zap.Error(err))
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
