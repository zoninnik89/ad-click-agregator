package main

import (
	"context"
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

	adId, _ := primitive.ObjectIDFromHex(id)

	var ad Ad
	err := collection.FindOne(ctx, bson.M{
		"_id":          adId,
		"advertiserId": advertiserId,
	}).Decode(&ad)

	return &ad, err
}
