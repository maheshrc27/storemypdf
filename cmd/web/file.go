package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/maheshrc27/storemypdf/internal/database"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

func (app *application) UploadFile(w http.ResponseWriter, r *http.Request) {
	workDir, _ := os.Getwd()

	err := r.ParseMultipartForm(15 << 20)
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

	err = app.db.InsertFile(fileData.ID, fileData.FileName, fileData.Description, fileData.FileType, fileData.Size, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app *application) DeleteFile(w http.ResponseWriter, r *http.Request) {
	workDir, _ := os.Getwd()

	fileid := r.URL.Query().Get("id")

	_, found, err := app.db.GetFile(fileid)

	if !found {
		http.Error(w, "File Not Found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = app.db.DeleteFile(fileid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = os.RemoveAll(filepath.Join(workDir, "uploads", fileid))
	if err != nil {
		fmt.Println(err)
	}
}
