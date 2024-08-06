package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/go-chi/chi/v5"
	"github.com/maheshrc27/storemypdf/internal/database"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

func (app *application) UploadFileApi(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	workDir, _ := os.Getwd()

	err = r.ParseMultipartForm(15 << 20)
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
	dst, err := os.Create(filepath.Join(uploadDir, fmt.Sprintf("%s%s", fileData.ID, ext)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	fileData.Size = int64(header.Size)
	written, err := io.Copy(dst, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println(written)

	// Detect the MIME type of the uploaded file
	mtype, err := mimetype.DetectFile(filepath.Join(uploadDir, fmt.Sprintf("%s%s", fileData.ID, ext)))
	if err != nil {
		fmt.Println("couldn't get file type")
		fileData.FileType = ""
	} else {
		fileData.FileType = mtype.String()
	}

	id, err = app.db.InsertFile(fileData.ID, fileData.FileName, fileData.Description, fileData.FileType, fileData.Size, userIDInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"success": true,
		"file_id": id,
		"message": "File uploaded successfully.",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func (app *application) FileInfoApi(w http.ResponseWriter, r *http.Request) {
	fileId := chi.URLParam(r, "file_id")

	fileInfo, found, err := app.db.GetFile(fileId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !found {
		http.Error(w, "Page Not Found", http.StatusNotFound)
		return
	}

	fileSize := float64(fileInfo.Size) / (1024 * 1024)

	data := map[string]interface{}{
		"success": true,
		"file": map[string]interface{}{
			"id":          fileInfo.ID,
			"name":        fileInfo.FileName,
			"type":        fileInfo.FileType,
			"size":        fileSize,
			"url":         "http://localhost:4444/f/" + fileInfo.ID,
			"description": fileInfo.Description,
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

}

func (app *application) FileDownloadApi(w http.ResponseWriter, r *http.Request) {
	workDir, _ := os.Getwd()
	fileId := chi.URLParam(r, "file_id")

	fileInfo, found, err := app.db.GetFile(fileId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !found {
		http.Error(w, "Page Not Found", http.StatusNotFound)
		return
	}

	filepath := filepath.Join(workDir, "uploads", fileId+".pdf")

	w.Header().Set("Content-Disposition", "attachment; filename="+fileInfo.FileName)
	http.ServeFile(w, r, filepath)
}

func (app *application) FileDeleteApi(w http.ResponseWriter, r *http.Request) {
	workDir, _ := os.Getwd()

	fileid := r.URL.Query().Get("file_id")

	_, found, err := app.db.GetFile(fileid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !found {
		http.Error(w, "File Not Found", http.StatusNotFound)
		return
	}

	err = app.db.DeleteFile(fileid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = os.RemoveAll(filepath.Join(workDir, "uploads", fileid))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
