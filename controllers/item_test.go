package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"lab1/collections"
	"lab1/database"
	"net/http"
	"net/http/httptest"
	"testing"
)

type ItemResponse struct {
	Data collections.Item `json:"data"`
}

func initDB() {
	if err := database.InitMongo("mongodb://localhost:27017", "test_DB"); err != nil {
		fmt.Println("errorInitDB", err)
	}
}

func TestCreateItem(t *testing.T) {
	// Create a test Gin router and recorder
	router := gin.Default()
	recorder := httptest.NewRecorder()
	router.POST("/item", CreateItem)

	// Create a test item
	item := collections.Item{Title: "Test Create Item", Status: true}

	// Convert item to JSON
	jsonItem, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}

	// Create a POST request with the JSON
	req, err := http.NewRequest("POST", "/item", bytes.NewBuffer(jsonItem))
	if err != nil {
		t.Fatal(err)
	}

	//initDB
	initDB()
	defer database.CloseMongoDB()

	// Perform the request
	router.ServeHTTP(recorder, req)

	// Check the response status code
	assert.Equal(t, http.StatusCreated, recorder.Code)

	// Check the response body
	var responseItem ItemResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &responseItem)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, item.Title, responseItem.Data.Title)
	assert.Equal(t, item.Status, responseItem.Data.Status)
}

func TestListItems(t *testing.T) {
	router := gin.Default()
	router.GET("/items", ListItems)
	recorder := httptest.NewRecorder()

	// Tạo request
	req, err := http.NewRequest("GET", "/items", nil)
	if err != nil {
		t.Fatal(err)
	}

	//Init mongo
	initDB()
	defer database.CloseMongoDB()

	router.ServeHTTP(recorder, req)

	// Check the response status code
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Check định dạng của response body
	expectedContentType := "application/json; charset=utf-8"
	if contentType := recorder.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Trả về định dạng Content-Type không đúng: expected %v but got %v", expectedContentType, contentType)
	}
}

func TestUpdateItem(t *testing.T) {
	var (
		router      = gin.Default()
		recorder    = httptest.NewRecorder()
		ctx, cancel = context.WithTimeout(context.Background(), database.CTimeOut)
	)
	defer cancel()

	//Init mongo
	initDB()
	defer database.CloseMongoDB()
	db := database.GetMongoDB()

	router.PUT("/item/:id", UpdateItem)

	// Insert a test item into the database
	item := collections.Item{ID: primitive.NewObjectID(), Title: "Test Create Item Update", Status: false}
	result, err := db.Collection("items").InsertOne(ctx, item)
	if err != nil {
		t.Fatal(err)
	}
	itemID := result.InsertedID.(primitive.ObjectID).Hex()

	// Create an updated test item
	updatedItem := collections.Item{ID: item.ID, Title: "Test Update Item", Status: true}

	// Convert updated item to JSON
	jsonUpdatedItem, err := json.Marshal(updatedItem)
	if err != nil {
		t.Fatal(err)
	}

	// Create a PUT request with the JSON updated item as the request body
	req, err := http.NewRequest("PUT", "/item/"+itemID, bytes.NewBuffer(jsonUpdatedItem))
	if err != nil {
		t.Fatal(err)
	}

	// Perform the request
	router.ServeHTTP(recorder, req)

	// Check the response status code
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Check the response body
	var responseItem ItemResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &responseItem)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, updatedItem.Title, responseItem.Data.Title)
	assert.Equal(t, updatedItem.Status, responseItem.Data.Status)

	// Check that the item was updated in the database
	var dbItem collections.Item
	err = db.Collection("items").FindOne(ctx, bson.M{"_id": result.InsertedID}).Decode(&dbItem)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, updatedItem.Title, dbItem.Title)
	assert.Equal(t, updatedItem.Status, dbItem.Status)
}

func TestChangeStatusItems(t *testing.T) {
	var (
		arrItemID   []primitive.ObjectID
		router      = gin.Default()
		recorder    = httptest.NewRecorder()
		ctx, cancel = context.WithTimeout(context.Background(), database.CTimeOut)
	)
	defer cancel()

	//Init mongo
	initDB()
	defer database.CloseMongoDB()
	db := database.GetMongoDB()

	router.POST("/change-status-items", ChangeStatusItems)

	// Insert a test item into the database
	for i := 0; i < 3; i++ {
		item := collections.Item{ID: primitive.NewObjectID(), Title: "Test Create Item Change Status", Status: false}
		result, err := db.Collection("items").InsertOne(ctx, item)
		if err != nil {
			t.Fatal(err)
		}
		itemID := result.InsertedID.(primitive.ObjectID)
		arrItemID = append(arrItemID, itemID)
	}

	data := ListIDRequest{ID: arrItemID}

	// Convert item to JSON
	jsonItem, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	// Create request
	req, err := http.NewRequest("POST", "/change-status-items", bytes.NewBuffer(jsonItem))
	if err != nil {
		t.Fatal(err)
	}

	// Perform the request
	router.ServeHTTP(recorder, req)

	// Check the response status code
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestDeleteItems(t *testing.T) {
	var (
		arrItemID   []primitive.ObjectID
		router      = gin.Default()
		recorder    = httptest.NewRecorder()
		ctx, cancel = context.WithTimeout(context.Background(), database.CTimeOut)
	)
	defer cancel()

	//Init mongo
	initDB()
	defer database.CloseMongoDB()
	db := database.GetMongoDB()

	router.POST("/delete-items", DeleteItems)

	// Insert items into the database
	for i := 0; i < 3; i++ {
		item := collections.Item{ID: primitive.NewObjectID(), Title: "Test Create Item Delete", Status: false}
		result, err := db.Collection("items").InsertOne(ctx, item)
		if err != nil {
			t.Fatal(err)
		}
		itemID := result.InsertedID.(primitive.ObjectID)
		arrItemID = append(arrItemID, itemID)
	}
	data := ListIDRequest{ID: arrItemID}

	// Convert item to JSON
	jsonItem, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	// Create a DELETE request
	req, err := http.NewRequest("POST", "/delete-items", bytes.NewBuffer(jsonItem))
	if err != nil {
		t.Fatal(err)
	}

	// Perform the request
	router.ServeHTTP(recorder, req)

	// Check the response status code
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Check that the item was deleted from the database
	count, err := db.Collection("items").CountDocuments(ctx, bson.M{
		"_id": bson.M{
			"$in": arrItemID,
		},
		"deleted_at": nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(0), count)
}
