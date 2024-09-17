package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	_ "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	_ "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	_ "go.mongodb.org/mongo-driver/mongo"
)

const (
	DbName   = "ads"
	CollName = "ads"
)

type Store struct {
	db *mongo.Client
}

func NewStore(db *mongo.Client) *Store {
	return &Store{db}
}

func (store *Store) Create(ctx context.Context, ad Ad) (primitive.ObjectID, error) {
	collection := store.db.Database(DbName).Collection(CollName)

	newAd, err := collection.InsertOne(ctx, ad)
	if err != nil {
		return primitive.NilObjectID, err
	}

	id := newAd.InsertedID.(primitive.ObjectID)
	return id, err

}

func (store *Store) Get(ctx context.Context, id, advertiserId string) (*Ad, error) {
	collection := store.db.Database(DbName).Collection(CollName)

	adId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ad ID: %v", err)
	}

	filter := bson.M{
		"_id":          adId,
		"advertiserID": advertiserId,
	}

	var ad Ad
	err = collection.FindOne(ctx, filter).Decode(&ad)

	return &ad, err
}
