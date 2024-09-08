package main

import (
	"context"
	"log"
	"net"

	common "github.com/zoninnik89/commons"

	"google.golang.org/grpc"
)

var (
	grpcAddr = common.EnvString("GRPC_ADDR", "localhost:2000")
)

func main() {

	grpcServer := grpc.NewServer()

	listner, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer listner.Close()

	store := NewStore()
	service := NewService(store)
	NewGrpcHandler(grpcServer)

	service.CreateAd(context.Background())

	log.Println("GRPC server started at ", grpcAddr)

	if err := grpcServer.Serve(listner); err != nil {
		log.Fatal(err.Error())
	}

}
