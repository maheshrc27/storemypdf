package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/maheshrc27/storemypdf/internal/database"
	"github.com/maheshrc27/storemypdf/internal/response"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

func (app *application) UploadFileApi(w http.ResponseWriter, r *http.Request) {

	userId, err := ParseUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	workDir, _ := os.Getwd()

	err = r.ParseMultipartForm(64 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	var fileData database.File
	fileData.FileName = header.Filename
	fileData.Description = r.FormValue("description")

	id, err := gonanoid.New(12)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fileData.ID = id

	if r.Form.Has("delete_after") {
		hrs := r.Form.Get("delete_after")
		currTime := time.Now()
		duration, _ := time.ParseDuration(fmt.Sprintf("%sh", hrs))
		deleteTime := currTime.Add(duration)
		app.db.InsertToDelete(id, deleteTime)
	}

	// Ensure the uploads directory exists
	uploadDir := filepath.Join(workDir, "uploads")
	if err := os.MkdirAll(uploadDir, 0750); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ext := filepath.Ext(header.Filename)

	if err := SaveFileToS3(fileData.ID, header.Filename, file); err != nil {
		handleError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fileData.Size = int64(header.Size)

	// Detect the MIME type of the uploaded file
	mtype, err := mimetype.DetectFile(filepath.Join(uploadDir, fmt.Sprintf("%s%s", fileData.ID, ext)))
	if err != nil {
		fmt.Println("couldn't get file type")
		fileData.FileType = ""
	} else {
		fileData.FileType = mtype.String()
	}

	id, err = app.db.InsertFile(fileData.ID, fileData.FileName, fileData.Description, fileData.FileType, fileData.Size, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"success":    true,
		"file_id":    id,
		"url_viewer": "storemypdf.com/f/" + id,
		"url":        "files.storemypadf.com/" + id + ".pdf",
		"message":    "File uploaded successfully",
	}

	err = response.JSONWithHeaders(w, 200, data, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app *application) GenerateApiKey(w http.ResponseWriter, r *http.Request) {

	userID, err := ParseUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	key := uuid.New().String()

	_, err = app.db.InsertKey(string(key), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Refresh", "true")
}

func (app *application) DeleteApiKey(w http.ResponseWriter, r *http.Request) {

	kid := r.URL.Query().Get("kid")

	keyid, err := strconv.Atoi(kid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = app.db.DeleteKey(int8(keyid))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Refresh", "true")
}

// func (app *application) FileInfoApi(w http.ResponseWriter, r *http.Request) {
// 	fileId := chi.URLParam(r, "file_id")

// 	fileInfo, found, err := app.db.GetFile(fileId)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	if !found {
// 		http.Error(w, "Page Not Found", http.StatusNotFound)
// 		return
// 	}

// 	fileSize := float64(fileInfo.Size) / (1024 * 1024)

// 	data := map[string]interface{}{
// 		"success": true,
// 		"file": map[string]interface{}{
// 			"id":          fileInfo.ID,
// 			"name":        fileInfo.FileName,
// 			"type":        fileInfo.FileType,
// 			"size":        fileSize,
// 			"url":         "http://localhost:4444/f/" + fileInfo.ID,
// 			"description": fileInfo.Description,
// 		},
// 	}

// 	err = response.JSONWithHeaders(w, 200, data, nil)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// }

// func (app *application) FileDeleteApi(w http.ResponseWriter, r *http.Request) {
// 	workDir, _ := os.Getwd()

// 	fileid := r.URL.Query().Get("file_id")

// 	file, found, err := app.db.GetFile(fileid)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	if !found {
// 		http.Error(w, "File Not Found", http.StatusNotFound)
// 		return
// 	}

// 	userId, err := ParseUserID(r)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	var data map[string]interface{}

// 	if file.UserID != userId {
// 		data = map[string]interface{}{
// 			"success": false,
// 			"message": "File doesn't belong to you.",
// 		}

// 		err = response.JSONWithHeaders(w, 400, data, nil)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}

// 	}

// 	err = app.db.DeleteFile(fileid)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 	}

// 	err = os.RemoveAll(filepath.Join(workDir, "uploads", fileid))
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	data = map[string]interface{}{
// 		"success": true,
// 		"message": "File deleted successfully.",
// 	}

// 	err = response.JSONWithHeaders(w, 200, data, nil)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }
