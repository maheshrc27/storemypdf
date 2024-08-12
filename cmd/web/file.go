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
	"github.com/maheshrc27/storemypdf/internal/funcs"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	premiumFileSizeLimit        = 50 << 20 // 50 MB
	standardFileSizeLimit int64 = 15 << 20 // 15 MB
	uploadDirPermissions        = 0750
)

func (app *application) UploadFile(w http.ResponseWriter, r *http.Request) {
	uid := r.Header.Get("X-User-ID")
	userId, err := getUserID(uid, app.db)
	if err != nil {
		handleError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	isPremium, err := funcs.CheckPremium(userId, app.db)
	if err != nil {
		handleError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	maxFileSize := standardFileSizeLimit
	if isPremium {
		maxFileSize = premiumFileSizeLimit
	}

	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		handleError(w, err.Error(), http.StatusForbidden)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		handleError(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileData := database.File{
		FileName:    header.Filename,
		Description: r.FormValue("description"),
		UserID:      userId,
		Size:        header.Size,
	}

	fileData.ID, err = gonanoid.New(12)
	if err != nil {
		handleError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uploadDir := filepath.Join(getWorkingDir(), "uploads")
	if err := os.MkdirAll(uploadDir, uploadDirPermissions); err != nil {
		handleError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := saveFile(uploadDir, fileData.ID, header.Filename, file); err != nil {
		handleError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fileData.FileType, err = detectFileType(filepath.Join(uploadDir, fileData.ID+filepath.Ext(header.Filename)))
	if err != nil {
		fileData.FileType = ""
	}

	if _, err := app.db.InsertFile(fileData.ID, fileData.FileName, fileData.Description, fileData.FileType, fileData.Size, fileData.UserID); err != nil {
		handleError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if deleteAfter := r.FormValue("delete_after"); deleteAfter != "" {
		if err := handleDeletionTime(deleteAfter, fileData.ID, app.db); err != nil {
			handleError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/f/"+fileData.ID, http.StatusSeeOther)
}

func (app *application) DeleteFile(w http.ResponseWriter, r *http.Request) {
	workDir, err := os.Getwd()
	if err != nil {
		http.Error(w, "Failed to determine working directory", http.StatusInternalServerError)
		return
	}

	fileID := r.URL.Query().Get("id")
	if fileID == "" {
		http.Error(w, "File ID is required", http.StatusBadRequest)
		return
	}

	_, found, err := app.db.GetFile(fileID)
	if err != nil {
		http.Error(w, "Error checking file existence", http.StatusInternalServerError)
		return
	}
	if !found {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	if err := app.db.DeleteFile(fileID); err != nil {
		http.Error(w, "Error deleting file record", http.StatusInternalServerError)
		return
	}

	filePath := filepath.Join(workDir, "uploads", fileID)
	if err := os.RemoveAll(filePath); err != nil {
		fmt.Printf("Error removing file from filesystem: %v\n", err)
		http.Error(w, "Error removing file from filesystem", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File successfully deleted"))
}

func getUserID(uid string, db *database.DB) (uuid.UUID, error) {
	if uid != "" {
		return uuid.Parse(uid)
	}
	return funcs.GuestId(db)
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
