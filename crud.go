package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gorilla/mux"
)

// Define a struct to represent the data model for our items in DynamoDB
type Item struct {
	ID  string ``
	Name string ``
}

// Initialize the DynamoDB client
var svc = dynamodb.New(session.Must(session.NewSessionWithOptions(session.Options{
	SharedConfigState: session.SharedConfigEnable,
})))

func main() {
	router := mux.NewRouter()

	// Define the routes for our API
	router.HandleFunc("/items", getAllItems).Methods("GET")
	router.HandleFunc("/items/{id}", getItem).Methods("GET")
	router.HandleFunc("/items", createItem).Methods("POST")
	router.HandleFunc("/items/{id}", updateItem).Methods("PUT")
	//router.HandleFunc("/items/{id}", deleteItem).Methods("DELETE")

	// Start the server
	fmt.Println("Server listening on port 8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}

// Get all items
func getAllItems(w http.ResponseWriter, r *http.Request) {
	// Define the name of your DynamoDB table
	tableName := "test"

	// Create the input for the Scan operation
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	// Call the Scan operation
	scanOutput, err := svc.Scan(input)
	if err != nil {
		log.Fatalf("Failed to scan items: %v", err)
	}

	// Unmarshal the items from the ScanOutput into a slice of Item structs
	items := []Item{}
	err = dynamodbattribute.UnmarshalListOfMaps(scanOutput.Items, &items)
	if err != nil {
		log.Fatalf("Failed to unmarshal items: %v", err)
	}

	// Write the items as a JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// Get an item by ID
func getItem(w http.ResponseWriter, r *http.Request) {
	// Define the name of your DynamoDB table
	tableName := "test"

	// Get the ID parameter from the request URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Create the input for the GetItem operation
	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(id),
			},
		},
	}

	// Call the GetItem operation
	getOutput, err := svc.GetItem(input)
	if err != nil {
		log.Fatalf("Failed to get item with ID %s: %v", id, err)
	}

	// Check if the item exists
	if len(getOutput.Item) == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Item with ID %s not found", id)
		return
	}

	// Unmarshal the item from the GetItemOutput into an Item struct
	item := Item{}
	err = dynamodbattribute.UnmarshalMap(getOutput.Item, &item)
	if err != nil {
		log.Fatalf("Failed to unmarshal item: %v", err)
	}

	// Write the item as a JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}
// Create a new item
func createItem(w http.ResponseWriter, r *http.Request) {
	// Define the name of your DynamoDB table
	tableName := "test"

	// Parse the request body into an Item struct
	item := Item{}
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Failed to parse request body: %v", err)
		return
	}

	// Marshal the Item struct into a DynamoDB attribute value map
	itemMap, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		log.Fatalf("Failed to marshal item: %v", err)
	}

	// Create the input for the PutItem operation
	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      itemMap,
	}
    log.Println("Input : ", input)
	// Call the PutItem operation
	_, err = svc.PutItem(input)
	if err != nil {
		log.Fatalf("Failed to create item: %v", err)
	}

	// Write the created item as a JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// Update an item by ID
func updateItem(w http.ResponseWriter, r *http.Request) {
	// Define the name of your DynamoDB table
	tableName := "test"

	// Get the ID parameter from the request URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Parse the request body into an Item struct
	item := Item{}
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Failed to parse request body: %v", err)
		return
	}

	// Marshal the Item struct into a DynamoDB attribute value map
	//itemMap, err := dynamodbattribute.MarshalMap(item)
	//if err != nil {
	//	log.Fatalf("Failed to marshal item: %v", err)
	//}

	// Create the input for the UpdateItem operation
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(id),
			},
		},
		UpdateExpression: aws.String("SET #name = :name"),
		ExpressionAttributeNames: map[string]*string{
			"#name": aws.String("Name"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":name": {
				S: aws.String(item.Name),
			},
		},
	}

	// Call the UpdateItem operation
	_, err = svc.UpdateItem(input)
	if err != nil {
		log.Fatalf("Failed to update item with ID %s: %v", id, err)
	}

	// Write the updated item as a JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}