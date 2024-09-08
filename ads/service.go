package main

import "context"

type Service struct {
	store AdsStore
}

func NewService(store AdsStore) *Service {
	return &Service{store}
}

func (s *Service) CreateAd(context.Context) error {
	return nil
}
