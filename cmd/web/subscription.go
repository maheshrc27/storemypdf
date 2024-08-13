package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/maheshrc27/storemypdf/internal/database"
	"github.com/maheshrc27/storemypdf/internal/paddle"
)

func (app *application) PaddleWebhook(w http.ResponseWriter, r *http.Request) {
	signature := r.Header.Get("Paddle-Signature")
	if signature == "" {
		http.Error(w, "Missing Paddle-Signature header", http.StatusUnauthorized)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	isValid, err := paddle.ValidateSignature(signature, string(body), os.Getenv("PADDLE_WEBHOOK_SECRET"))
	if err != nil {
		fmt.Println("error occured")
		http.Error(w, "Invalid webhook signature", http.StatusUnauthorized)
	}
	if !isValid {
		fmt.Println("not valid")
		http.Error(w, "Invalid webhook signature", http.StatusUnauthorized)
		return
	}

	var parsedBody map[string]interface{}
	if err := json.Unmarshal(body, &parsedBody); err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	eventType, ok := parsedBody["event_type"].(string)
	if !ok {
		http.Error(w, "Invalid event type", http.StatusBadRequest)
		return
	}

	handleEvent(app.db, eventType, parsedBody, w)
}

func handleEvent(db *database.DB, eventType string, parsedBody map[string]interface{}, w http.ResponseWriter) {
	data, ok := parsedBody["data"].(map[string]interface{})
	if !ok {
		log.Printf("data is not a map")
	}

	switch eventType {
	case "subscription.created":
		handleSubscriptionCreated(db, data, w)
	case "subscription.updated":
		handleSubscriptionUpdated(db, data, w)
	case "subscription.canceled":
		handleSubscriptionCanceled(db, data, w)
	case "subscription.past_due":
		handleSubscriptionPastDue(db, data, w)
	default:
		http.Error(w, "Unknown event type", http.StatusBadRequest)
	}
}

func handleSubscriptionCreated(db *database.DB, data map[string]interface{}, w http.ResponseWriter) {

	items, ok := data["items"].([]interface{})
	if !ok {
		log.Printf("items is not a slice")
	}

	var productId string
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			log.Printf("item is not a map")
		}
		price, ok := itemMap["price"].(map[string]interface{})
		if !ok {
			log.Printf("price is not a map")
		}

		productID, ok := price["product_id"].(string)
		if !ok {
			log.Printf("price.product_id is not a string")
		}
		productId = productID
	}

	subscription_id, ok := data["id"].(string)
	if !ok {
		log.Printf("id is not a string")
	}
	status, ok := data["status"].(string)
	if !ok {
		log.Printf("status is not a string")
	}
	nextBill, ok := data["next_billed_at"].(string)
	if !ok {
		log.Printf("next_billed_at is not a string")
	}
	parsedNextBill, err := time.Parse(time.RFC3339, nextBill)
	if err != nil {
		log.Printf("Time couldn't be parsed: %v", err)
	}

	customData, ok := data["custom_data"].(map[string]interface{})
	if !ok {
		log.Printf("customData is not a string")
	}

	fmt.Println(customData)

	uid, ok := customData["user_id"].(string)
	if !ok {
		log.Printf("user_id is not a string")
	}

	userId, err := uuid.Parse(uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = db.InsertSubscription(subscription_id, productId, status, parsedNextBill, userId)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Subscription created successfully")
}

func handleSubscriptionUpdated(db *database.DB, data map[string]interface{}, w http.ResponseWriter) {
	subscription_id, ok := data["id"].(string)
	if !ok {
		log.Printf("id is not a string")
	}
	status, ok := data["status"].(string)
	if !ok {
		log.Printf("status is not a string")
	}
	nextBill, ok := data["next_billed_at"]
	if !ok {
		log.Printf("next_billed_at is not a string")
	}

	if nextBill == nil {
		err := db.UpdateSubscriptionStatus(status, subscription_id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		parsedNextBill, err := time.Parse(time.RFC3339, nextBill.(string))
		if err != nil {
			log.Printf("Time couldn't be parsed: %v", err)
		}

		err = db.UpdateSubscriptionStatusAndNextBill(status, parsedNextBill, subscription_id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Subscription updated successfully")
}

func handleSubscriptionCanceled(db *database.DB, data map[string]interface{}, w http.ResponseWriter) {
	subscription_id, ok := data["id"].(string)
	if !ok {
		log.Printf("id is not a string")
	}
	status, ok := data["status"].(string)
	fmt.Println(status)
	if !ok {
		log.Printf("status is not a string")
	}

	err := db.UpdateSubscriptionStatus(status, subscription_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Subscription canceled successfully")
}

func handleSubscriptionPastDue(db *database.DB, data map[string]interface{}, w http.ResponseWriter) {
	subscription_id, ok := data["id"].(string)
	if !ok {
		log.Printf("id is not a string")
	}

	status, ok := data["status"].(string)
	if !ok {
		log.Printf("status is not a string")
	}

	err := db.UpdateSubscriptionStatus(status, subscription_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Subscription past due handled successfully")
}
