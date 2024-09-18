package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/zoninnik89/ad-click-aggregator/gateway/gateway"
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

	registry, err := consul.NewRegistry(consulAddress, serviceName)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, httpAddress); err != nil {
		panic(err)
	}

	go func() {
		for {
			if err := registry.HealthCheck(instanceID, serviceName); err != nil {
				log.Fatal("failed to health check")
			}
			time.Sleep(time.Second * 1)
		}
	}()

	defer registry.Deregister(ctx, instanceID, serviceName)

	producer, err := NewProducer()
	if err != nil {
		log.Fatalf("failed to initialize producer: %v", err)
	}

	defer producer.Close()

	fmt.Printf("Kafka PRODUCER 📨 started at http://localhost%s\n", ProducerPort)

	if err := router.Run(ProducerPort); err != nil {
		log.Printf("failed to run the server: %v", err)

	mux := http.NewServeMux()
	adsGateway := gateway.NewGRPCGateway(registry)

	handler := NewHandler(adsGateway)
	handler.registerRoutes(mux)

	log.Printf("Starting HTTP server at %s", httpAddress)

	if err := http.ListenAndServe(httpAddress, mux); err != nil {
		log.Fatal("Failed to start http server")
	}
}
