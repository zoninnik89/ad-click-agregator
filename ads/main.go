package main

import "context"

func main() {

	store := NewStore()
	service := NewService(store)

	service.CreateAd(context.Background())
}
