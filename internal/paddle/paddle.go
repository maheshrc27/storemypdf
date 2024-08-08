package paddle

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
)

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

func ValidateSignature(signature, body, secret string) (bool, error) {
	ts, h1, err := extractValues(signature)
	if err != nil {
		return false, err
	}

	return hashSignature(ts, body, h1, secret)
}
