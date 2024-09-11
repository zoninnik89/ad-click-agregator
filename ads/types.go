package main

import (
	"context"

	protoBuff "github.com/zoninnik89/commons/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AdsService interface {
	GetAd(context.Context, *protoBuff.GetAdRequest) (*protoBuff.Ad, error)
	CreateAd(context.Context, *protoBuff.CreateAdRequest) (*protoBuff.Ad, error)
}

type AdsStore interface {
	Get(ctx context.Context, adId, advertiserID string) (*Ad, error)
	Create(context.Context, Ad) (primitive.ObjectID, error)
}

type Ad struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	AdvertiserID string             `bson:"advertiserID,omitempty"`
	Title        string             `bson:"title,omitempty"`
	AdURL        string             `bson:"adURL,omitempty"`
}

func (ad *Ad) ToProto() *protoBuff.Ad {
	return &protoBuff.Ad{
		ID:           ad.ID.Hex(),
		AdvertiserID: ad.AdvertiserID,
		Title:        ad.Title,
		AdURL:        ad.AdURL,
	}
}
