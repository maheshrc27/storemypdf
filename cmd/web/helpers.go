package main

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/maheshrc27/storemypdf/internal/database"
	"github.com/maheshrc27/storemypdf/internal/version"
)

func (app *application) newTemplateData(templateData interface{}) map[string]any {
	data := map[string]any{
		"Version": version.Get(),
	}
	if templateDataMap, ok := templateData.(map[string]any); ok {
		for key, value := range templateDataMap {
			data[key] = value
		}
	}

	return data
}

func (app *application) newEmailData() map[string]any {
	data := map[string]any{
		"BaseURL": app.config.baseURL,
	}

	return data
}

func (app *application) backgroundTask(r *http.Request, fn func() error) {
	app.wg.Add(1)

	go func() {
		defer app.wg.Done()

		defer func() {
			err := recover()
			if err != nil {
				app.reportServerError(r, fmt.Errorf("%s", err))
			}
		}()

		err := fn()
		if err != nil {
			app.reportServerError(r, err)
		}
	}()
}

func getUserID(uid string, db *database.DB) (uuid.UUID, error) {
	if uid != "" {
		return uuid.Parse(uid)
	}
	return GuestId(db)
}

func handleDeletionTime(deleteAfter string, fileID string, db *database.DB) error {
	hrs, err := time.ParseDuration(deleteAfter + "h")
	if err != nil {
		return fmt.Errorf("invalid delete_after duration: %v", err)
	}
	deleteTime := time.Now().Add(hrs)

	_, err = db.InsertToDelete(fileID, deleteTime)
	if err != nil {
		return err
	}

	return nil
}

func saveFile(uploadDir, fileID, filename string, file multipart.File) error {
	ext := filepath.Ext(filename)
	dst, err := os.Create(filepath.Join(uploadDir, fileID+ext))
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, file)
	return err
}

func detectFileType(filePath string) (string, error) {
	mtype, err := mimetype.DetectFile(filePath)
	if err != nil {
		return "", err
	}
	return mtype.String(), nil
}

func getWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Failed to get working directory: %v", err))
	}
	return dir
}

func GetAuthStatus(r *http.Request) bool {
	isLoggedIn := r.Header.Get("X-Logged-IN")
	loggedIn := false

	if isLoggedIn == "true" {
		loggedIn = true
	}

	return loggedIn
}

func ParseUserID(r *http.Request) (uuid.UUID, error) {

	uid := r.Header.Get("X-User-ID")

	userId, err := uuid.Parse(uid)
	if err != nil {
		return uuid.Nil, err
	}
	return userId, nil
}

func GuestId(db *database.DB) (uuid.UUID, error) {
	var user *database.User
	user, _, err := db.GetUserByEmail("guest@storemypdf.com")
	if err != nil {
		return uuid.Nil, err
	}
	return user.ID, nil
}

func CheckPremium(userId uuid.UUID, db *database.DB) (*database.Subscription, bool, error) {
	subscription, found, err := db.GetSubscriptionByUserID(userId)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}

	return subscription, isActiveSubscription(subscription), nil
}

func isActiveSubscription(subscription *database.Subscription) bool {
	now := time.Now()
	switch subscription.Status {
	case "active":
		return true
	case "canceled", "past_due", "paused":
		return subscription.NextBillDate.After(now)
	default:
		return false
	}
}
