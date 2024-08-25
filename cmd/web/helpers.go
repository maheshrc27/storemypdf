package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/maheshrc27/storemypdf/internal/database"
	"github.com/maheshrc27/storemypdf/internal/response"
	"github.com/maheshrc27/storemypdf/internal/tokens"
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

func (app *application) SendVerificationEmail(w http.ResponseWriter, email string, userid uuid.UUID) {

	tokenString, err := tokens.GenerateJWT(userid.String(), app.config.cookie.secretKey)
	if err != nil {
		response.HTML(w, "Internal Server Error")
		return
	}

	data := map[string]any{
		"Name": "storemypdf",
		"link": "http://localhost:4444/verify-email?vid=" + tokenString,
	}

	err = app.mailer.Send(email, data, "verification.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func SaveFileToS3(fileID, filename string, file multipart.File) error {
	ext := filepath.Ext(filename)

	const (
		AWS_S3_REGION = ""
		AWS_S3_BUCKET = ""
	)

	session, err := session.NewSession(&aws.Config{Region: aws.String(AWS_S3_REGION)})
	if err != nil {
		log.Fatal(err)
	}

	uploader := s3manager.NewUploader(session)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(AWS_S3_BUCKET),
		Key:    aws.String(fileID + ext),
		Body:   file,
	})

	if err != nil {
		fmt.Printf("failed to upload file, %v", err)
		return nil
	}

	return nil
}

func DeleteS3Object(fileID string) error {
	const (
		AWS_S3_REGION = ""
		AWS_S3_BUCKET = ""
	)

	session, err := session.NewSession(&aws.Config{Region: aws.String(AWS_S3_REGION)})
	if err != nil {
		log.Fatal(err)
	}

	svc := s3.New(session)

	_, err = svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(AWS_S3_BUCKET),
		Key:    aws.String(fileID + ".pdf"),
	})

	if err != nil {
		fmt.Printf("failed to upload file, %v", err)
		return nil
	}

	return nil
}
