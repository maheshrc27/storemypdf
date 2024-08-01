package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/maheshrc27/storemypdf/internal/database"
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
	isValid, err := validateSignature(signature, string(body), os.Getenv("PADDLE_WEBHOOK_SECRET"))
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

	userId, ok := customData["userId"].(string)
	if !ok {
		log.Printf("userId is not a string")
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

func hashSignature(ts, requestBody, h1, secretKey string) (bool, error) {
	payload := fmt.Sprintf("%s:%s", ts, requestBody)
	key := []byte(secretKey)

	h := hmac.New(sha256.New, key)
	_, err := h.Write([]byte(payload))
	if err != nil {
		return false, err
	}

	signature := h.Sum(nil)
	signatureHex := hex.EncodeToString(signature)

	if !hmac.Equal([]byte(signatureHex), []byte(h1)) {
		return false, nil
	}
	return true, nil
}

func extractValues(input string) (string, string, error) {
	tsRegex := regexp.MustCompile(`ts=(\d+)`)
	h1Regex := regexp.MustCompile(`h1=([a-f0-9]+)`)

	matchTs := tsRegex.FindStringSubmatch(input)
	matchH1 := h1Regex.FindStringSubmatch(input)

	if len(matchTs) < 2 || len(matchH1) < 2 {
		return "", "", errors.New("invalid input format")
	}

	return matchTs[1], matchH1[1], nil
}

func validateSignature(signature, body, secret string) (bool, error) {
	ts, h1, err := extractValues(signature)
	if err != nil {
		return false, err
	}

	return hashSignature(ts, body, h1, secret)
}
