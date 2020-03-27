package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/DenisAltruist/distsys/db"
	"go.mongodb.org/mongo-driver/bson"
	mgo "go.mongodb.org/mongo-driver/mongo"
)

type clientResponse struct {
	Text string
	Code int
}

func sendBodyResponse(w http.ResponseWriter, text string, code int) {
	resp := clientResponse{
		Text: text,
		Code: code,
	}
	encodedJson, _ := json.Marshal(&resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, "%s\n", string(encodedJson))
}

func sendError(w http.ResponseWriter, code int, formatMsg string, args ...interface{}) {
	msg := formatMsg
	if len(args) != 0 {
		msg = fmt.Sprintf(formatMsg, args)
	}
	log.Printf(msg)
	sendBodyResponse(w, msg, code)
}

func getDBClient(w http.ResponseWriter) (*mgo.Client, bool) {
	mongoConnStr := os.Getenv("MONGO_CONN_STRING")
	client, err := db.CreateSession(mongoConnStr, 10*time.Second)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Can't connect to database, got an error: %s", err.Error())
		return nil, false
	}
	return client, true
}

func getItemFromRequest(w http.ResponseWriter, r *http.Request) (*db.StoreItem, bool) {
	contents, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Can't parse request body, got error: %s", err.Error())
		return nil, false
	}
	var newItem db.StoreItem
	err = json.Unmarshal(contents, &newItem)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Can't unrmashal contents %s, expected valid JSON", string(contents))
		return nil, false
	}
	return &newItem, true
}

func createItem(w http.ResponseWriter, r *http.Request) {
	newItem, ok := getItemFromRequest(w, r)
	if !ok {
		return
	}
	client, ok := getDBClient(w)
	if !ok {
		return
	}
	filter := bson.D{bson.E{Key: "code", Value: newItem.Code}} // Maintenance of uniqueness of codes
	isAlreadyAdded, err := db.DoesItemExist(client, &filter, 5*time.Second)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Can't check if there is another item with code %s", newItem.Code)
		return
	}
	if isAlreadyAdded {
		sendError(w, http.StatusBadRequest, "There is another item with code %s already created", newItem.Code)
		return
	}
	err = db.AddItem(client, newItem, 5*time.Second)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Can't add item, got an error: %s", err.Error())
		return
	}
	sendBodyResponse(w, "Success", http.StatusOK)
}

func showItem(w http.ResponseWriter, r *http.Request) {
	filterKey := "code"
	filterVal := r.FormValue(filterKey)
	if len(filterVal) == 0 {
		sendError(w, http.StatusBadRequest, "'%s' argument is not specified", filterKey)
		return
	}
	filter := bson.M{filterKey: filterVal}
	client, ok := getDBClient(w)
	if !ok {
		return
	}
	items, err := db.FindItems(client, &filter, 5*time.Second)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Can't find item with code %s, got an error: %s", filterVal, err.Error())
		return
	}
	if items.Count == 0 {
		sendError(w, http.StatusBadRequest, "There is no item with code %s to show", filterVal)
		return
	}
	encodedItem, err := json.Marshal(items.List[0])
	if err != nil {
		sendError(w, http.StatusBadRequest, "Can't marshal found item: %s", err.Error())
		return
	}
	log.Printf("Found item JSON: %s", encodedItem)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(encodedItem))
}

func removeItem(w http.ResponseWriter, r *http.Request) {
	filterKey := "code"
	filterVal := r.FormValue(filterKey)
	if len(filterVal) == 0 {
		sendError(w, http.StatusBadRequest, "'%s' argument is not specified", filterKey)
		return
	}
	filter := bson.M{filterKey: filterVal}
	client, ok := getDBClient(w)
	if !ok {
		return
	}
	removeCount, err := db.RemoveItem(client, &filter, 5*time.Second)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Can't remove item: %s", err.Error())
		return
	}
	if removeCount == 0 {
		sendError(w, http.StatusBadRequest, "There is no item with code %s to remove", filterVal)
		return
	}
	sendBodyResponse(w, "Success", http.StatusOK)
}

func showItemsList(w http.ResponseWriter, r *http.Request) {
	filterKey := "category"
	filterVal := r.FormValue(filterKey)
	if len(filterVal) == 0 {
		sendError(w, http.StatusBadRequest, "'%s' argument is not specified", filterKey)
		return
	}
	filter := bson.M{filterKey: filterVal}
	client, ok := getDBClient(w)
	if !ok {
		return
	}
	items, err := db.FindItems(client, &filter, 5*time.Second)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Can't remove item: %s", err.Error())
		return
	}
	log.Printf("Num of items: %d\n", items.Count)
	encodedItems, err := json.Marshal(items)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Can't marshal set of items to JSON: %s", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(encodedItems))
}

func editItem(w http.ResponseWriter, r *http.Request) {
	filterKey := "code"
	filterVal := r.FormValue(filterKey)
	if len(filterVal) == 0 {
		sendError(w, http.StatusBadRequest, "'%s' argument is not specified", filterKey)
		return
	}
	filter := bson.D{bson.E{Key: filterKey, Value: filterVal}}
	newItemFields, ok := getItemFromRequest(w, r)
	if !ok {
		return
	}
	client, ok := getDBClient(w)
	if !ok {
		return
	}
	err := db.UpdateItem(client, &filter, newItemFields, 5*time.Second)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Can't update item: %s", err.Error())
		return
	}
	sendBodyResponse(w, "Success", http.StatusOK)
}

func main() {
	http.HandleFunc("/create_item", createItem)
	http.HandleFunc("/show_item", showItem)
	http.HandleFunc("/remove_item", removeItem)
	http.HandleFunc("/show_items_list", showItemsList)
	http.HandleFunc("/edit_item", editItem)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("EXTERNAL_LISTEN_PORT")), nil))
}