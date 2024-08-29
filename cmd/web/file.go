package main

import (
	"net/http"

	"github.com/maheshrc27/storemypdf/internal/database"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	premiumFileSizeLimit        = 64 << 20 // 50 MB
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

	_, premium, err := CheckPremium(userId, app.db)
	if err != nil {
		handleError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	maxFileSize := standardFileSizeLimit
	if premium {
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

	if err := SaveFileToS3(fileData.ID, header.Filename, file); err != nil {
		handleError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fileData.FileType = ""

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

	err = DeleteS3Object(fileID)
	if err != nil {
		http.Error(w, "Error deleting file record", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Refresh", "true")
}
