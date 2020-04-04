package db

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	mgo "go.mongodb.org/mongo-driver/mongo"
)

type ShopUser struct {
	Email        string `bson:"email" json:"email"`
	Password     string `bson:"password" json:"password"`
	PasswordHash string `bson:"password_hash" json:"password_hash"`
}

type TokensPair struct {
	Email        string `bson:"email" json:"email"`
	AccessToken  string `bson:"access_token" json:"access_token"`
	RefreshToken string `bson:"refresh_token" json:"refresh_token"`
}

func AddNewUser(client *mgo.Client, user *ShopUser, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collection := getUsersCollection(client)
	insertRes, err := collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	log.Printf("Inserted doc id: %s", insertRes.InsertedID)
	return nil
}

func FindUser(client *mgo.Client, filter *bson.D, timeout time.Duration) (*ShopUser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collection := getUsersCollection(client)
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	isUserFound := cur.Next(context.Background())
	if cur.Err() != nil {
		return nil, cur.Err()
	}
	if !isUserFound {
		return nil, errors.New("Can't find such user in database")
	}
	var res ShopUser
	cur.Decode(&res)
	return &res, nil
}
