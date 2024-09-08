package main

import "context"

type AdsService interface {
	CreateAd(context.Context) error
}

type AdsStore interface {
	Create(context.Context) error
}
