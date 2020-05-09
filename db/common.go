package db

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/DenisAltruist/distsys/utils"
	"go.mongodb.org/mongo-driver/bson"
	mgo "go.mongodb.org/mongo-driver/mongo"
	mgopts "go.mongodb.org/mongo-driver/mongo/options"
)

func ToBsonDoc(v interface{}) (doc *bson.D, err error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return
	}
	err = bson.Unmarshal(data, &doc)
	return
}

func CreateSession(connStr string, timeout time.Duration) (*mgo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	client, err := mgo.Connect(ctx, mgopts.Client().ApplyURI(connStr))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func GetDbClient(w http.ResponseWriter) (*mgo.Client, bool) {
	mongoConnStr := os.Getenv("MONGO_CONN_STRING")
	client, err := CreateSession(mongoConnStr, 10*time.Second)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Can't connect to database, got an error: %s", err.Error())
		return nil, false
	}
	return client, true
}

func getItemsCollection(client *mgo.Client) *mgo.Collection {
	return client.Database(os.Getenv("MONGO_SHOP_DB_NAME")).Collection(os.Getenv("MONGO_ITEMS_COLL_NAME"))
}

func GetActiveUsersCollection(client *mgo.Client) *mgo.Collection {
	return client.Database(os.Getenv("MONGO_SHOP_DB_NAME")).Collection(os.Getenv("MONGO_ACTIVE_USERS_COLL_NAME"))
}

func GetPendingUsersCollection(client *mgo.Client) *mgo.Collection {
	return client.Database(os.Getenv("MONGO_SHOP_DB_NAME")).Collection(os.Getenv("MONGO_PENDING_USERS_COLL_NAME"))
}
