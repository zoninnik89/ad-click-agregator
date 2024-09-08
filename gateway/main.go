package main

import (
	"log"
	"net/http"

	_ "github.com/joho/godotenv/autoload"
	common "github.com/zoninnik89/commons"
	protoBuff "github.com/zoninnik89/commons/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	httpAddress    = common.EnvString("HTTP_ADDR", ":8080")
	adsServiceAddr = "localhost:2000"
)

func main() {
	conn, err := grpc.NewClient(adsServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to establish connection with server: %v", err)
	}
	defer conn.Close()

	log.Println("Dialing ads service at ", adsServiceAddr)

	client := protoBuff.NewAdsServiceClient(conn)

	mux := http.NewServeMux()
	handler := NewHandler(client)
	handler.registerRoutes(mux)

	log.Printf("Starting HTTP server at %s", httpAddress)

	if err := http.ListenAndServe(httpAddress, mux); err != nil {
		log.Fatal("Failed to start http server")
	}
}
