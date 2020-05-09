package db

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	mgo "go.mongodb.org/mongo-driver/mongo"
)

type NotifyRequest struct {
	Email   string `bson:"email" json:"email"`
	Message string `bson:"message" json:"message"`
}

type ShopUser struct {
	Email        string `bson:"email" json:"email"`
	Password     string `bson:"password" json:"password"`
	PasswordHash string `bson:"password_hash" json:"password_hash"`
	ConfirmToken string `bson:"confirm_token" json:"confirm_token"`
	CreatedAt    int64  `bson:"created_at" json:"created_at"`
}

type TokensPair struct {
	Email        string `bson:"email" json:"email"`
	AccessToken  string `bson:"access_token" json:"access_token"`
	RefreshToken string `bson:"refresh_token" json:"refresh_token"`
}

func RemoveFromPending(client *mgo.Client, user *ShopUser, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	pendingColl := GetPendingUsersCollection(client)
	filter := bson.D{bson.E{Key: "confirm_token", Value: user.ConfirmToken}}
	delRes, err := pendingColl.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if delRes.DeletedCount != 1 {
		return errors.New("Deleted count on pending user in confirmation is not 1")
	}
	return nil
}

func ConfirmUser(client *mgo.Client, user *ShopUser, timeout time.Duration) error {
	err := RemoveFromPending(client, user, timeout)
	if err != nil {
		return err
	}
	user.ConfirmToken = ""
	return AddNewUser(GetActiveUsersCollection(client), user, timeout)
}

// Collection for either pending/active users
func AddNewUser(coll *mgo.Collection, user *ShopUser, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	user.CreatedAt = time.Now().Unix()
	insertRes, err := coll.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	log.Printf("Inserted doc id: %s", insertRes.InsertedID)
	return nil
}

// Collection for either pending/active users
func FindUser(coll *mgo.Collection, filter *bson.D, timeout time.Duration) (*ShopUser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cur, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	isUserFound := cur.Next(context.Background())
	if cur.Err() != nil {
		return nil, cur.Err()
	}
	if !isUserFound {
		return nil, nil
	}
	var res ShopUser
	cur.Decode(&res)
	return &res, nil
}
