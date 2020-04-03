package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	mgo "go.mongodb.org/mongo-driver/mongo"
	mgopts "go.mongodb.org/mongo-driver/mongo/options"
)

type StoreItem struct {
	Name     string `bson:"name" json:"name"`
	Code     string `bson:"code" json:"code"`
	Category string `bson:"category" json:"category"`
}

type StoreItemsList struct {
	Count int64        `json:"count"`
	List  []*StoreItem `json:"list"`
}

func AddItem(client *mgo.Client, item *StoreItem, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collection := getItemsCollection(client)
	insertRes, err := collection.InsertOne(ctx, item)
	if err != nil {
		return err
	}
	log.Printf("Inserted doc id: %s", insertRes.InsertedID)
	return nil
}

func DoesItemExist(client *mgo.Client, filter *bson.D, timeout time.Duration) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collection := getItemsCollection(client)
	cur, err := collection.Find(ctx, *filter)
	if err != nil {
		return false, err
	}
	defer cur.Close(ctx)
	isItemFound := cur.Next(context.Background())
	if cur.Err() != nil {
		return false, cur.Err()
	}
	return isItemFound, nil
}

func FindItems(client *mgo.Client, filter *bson.M, offset int64, limit int64, timeout time.Duration) (*StoreItemsList, error) {
	var result StoreItemsList
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collection := getItemsCollection(client)
	cur, err := collection.Find(ctx, filter, &mgopts.FindOptions{
		Skip:  &offset,
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var curItem StoreItem
		err = cur.Decode(&curItem)
		if err != nil {
			return nil, err
		}
		result.List = append(result.List, &curItem)
	}
	numOfDocs, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}
	result.Count = numOfDocs
	return &result, nil
}

func RemoveItem(client *mgo.Client, filter *bson.M, timeout time.Duration) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collection := getItemsCollection(client)
	delRes, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return 0, err
	}
	log.Printf("Deleted documents count for %v: %d\n", filter, delRes.DeletedCount)
	return delRes.DeletedCount, nil
}

func UpdateItem(client *mgo.Client, filter *bson.D, newItemVal *StoreItem, timeout time.Duration) error {
	newItemBsonD, err := ToBsonDoc(newItemVal)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collection := getItemsCollection(client)
	updateRes, err := collection.UpdateOne(ctx, filter, bson.D{bson.E{Key: "$set", Value: newItemBsonD}})
	if err != nil {
		return err
	}
	log.Printf("Matched count: %d\n", updateRes.MatchedCount)
	if updateRes.MatchedCount != 1 {
		return errors.New(fmt.Sprintf("Can't match item %v", *filter))
	}
	return nil
}
