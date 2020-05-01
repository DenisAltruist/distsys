package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/DenisAltruist/distsys/db"
	"github.com/DenisAltruist/distsys/utils"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

func getItemFromRequest(w http.ResponseWriter, r *http.Request) (*db.StoreItem, bool) {
	contents, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Can't parse request body, got error: %s", err.Error())
		return nil, false
	}
	var newItem db.StoreItem
	err = json.Unmarshal(contents, &newItem)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Can't unrmashal contents %s, expected valid JSON", string(contents))
		return nil, false
	}
	return &newItem, true
}

func createItem(w http.ResponseWriter, r *http.Request) {
	newItem, ok := getItemFromRequest(w, r)
	if !ok {
		return
	}
	client, ok := db.GetDbClient(w)
	if !ok {
		return
	}
	filter := bson.D{bson.E{Key: "code", Value: newItem.Code}} // Maintenance of uniqueness of codes
	isAlreadyAdded, err := db.DoesItemExist(client, &filter, 5*time.Second)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Can't check if there is another item with code %s", newItem.Code)
		return
	}
	if isAlreadyAdded {
		utils.SendError(w, http.StatusBadRequest, "There is another item with code %s already created", newItem.Code)
		return
	}
	err = db.AddItem(client, newItem, 5*time.Second)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Can't add item, got an error: %s", err.Error())
		return
	}
	utils.SendBodyResponse(w, "Success", http.StatusOK)
}

func showItem(w http.ResponseWriter, r *http.Request) {
	filterKey := "code"
	filterVal := r.FormValue(filterKey)
	if len(filterVal) == 0 {
		utils.SendError(w, http.StatusBadRequest, "'%s' argument is not specified", filterKey)
		return
	}
	filter := bson.M{filterKey: filterVal}
	client, ok := db.GetDbClient(w)
	if !ok {
		return
	}
	items, err := db.FindItems(client, &filter, 0 /* offset */, 1 /* limit */, 5*time.Second)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Can't find item with code %s, got an error: %s", filterVal, err.Error())
		return
	}
	if items.Count == 0 {
		utils.SendError(w, http.StatusBadRequest, "There is no item with code %s to show", filterVal)
		return
	}
	encodedItem, err := json.Marshal(items.List[0])
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Can't marshal found item: %s", err.Error())
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
		utils.SendError(w, http.StatusBadRequest, "'%s' argument is not specified", filterKey)
		return
	}
	filter := bson.M{filterKey: filterVal}
	client, ok := db.GetDbClient(w)
	if !ok {
		return
	}
	removeCount, err := db.RemoveItem(client, &filter, 5*time.Second)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Can't remove item: %s", err.Error())
		return
	}
	if removeCount == 0 {
		utils.SendError(w, http.StatusBadRequest, "There is no item with code %s to remove", filterVal)
		return
	}
	utils.SendBodyResponse(w, "Success", http.StatusOK)
}

func showItemsList(w http.ResponseWriter, r *http.Request) {
	filterKey := "category"
	offsetStr := r.FormValue("offset") // pagination
	limitStr := r.FormValue("limit")   // pagination
	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Specified offset is not correct: %s", err.Error())
		return
	}
	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Specified limit is not correct: %s", err.Error())
		return
	}
	filterVal := r.FormValue(filterKey)
	if len(filterVal) == 0 {
		utils.SendError(w, http.StatusBadRequest, "'%s' argument is not specified", filterKey)
		return
	}
	filter := bson.M{filterKey: filterVal}
	client, ok := db.GetDbClient(w)
	if !ok {
		return
	}
	items, err := db.FindItems(client, &filter, offset, limit, 5*time.Second)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Can't remove item: %s", err.Error())
		return
	}
	log.Printf("Num of items: %d\n", items.Count)
	encodedItems, err := json.Marshal(items)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Can't marshal set of items to JSON: %s", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(encodedItems))
}

func editItem(w http.ResponseWriter, r *http.Request) {
	filterKey := "code"
	filterVal := r.FormValue(filterKey)
	if len(filterVal) == 0 {
		utils.SendError(w, http.StatusBadRequest, "'%s' argument is not specified", filterKey)
		return
	}
	filter := bson.D{bson.E{Key: filterKey, Value: filterVal}}
	newItemFields, ok := getItemFromRequest(w, r)
	if !ok {
		return
	}
	newItemFields.Code = filterVal // we forbid to change code of the requested item
	client, ok := db.GetDbClient(w)
	if !ok {
		return
	}
	err := db.UpdateItem(client, &filter, newItemFields, 5*time.Second)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Can't update item: %s", err.Error())
		return
	}
	utils.SendBodyResponse(w, "Success", http.StatusOK)
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/item", createItem).Methods("POST")
	router.HandleFunc("/item", removeItem).Methods("DELETE")
	router.HandleFunc("/item", showItem).Methods("GET")
	router.HandleFunc("/item", editItem).Methods("PUT")
	router.HandleFunc("/items", showItemsList).Methods("GET")
	router.Use(authMiddleware())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("EXTERNAL_LISTEN_PORT")), router))
}
