package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/maheshrc27/storemypdf/internal/response"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	err := response.Page(w, http.StatusOK, data, "pages/home.tmpl")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) ApiDocs(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	err := response.Page(w, http.StatusOK, data, "pages/docs.tmpl")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) FileInfo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fmt.Println("File ID:", id)

	// Fetch file info from the database
	fileInfo, found, err := app.db.GetFile(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !found {
		http.Error(w, "Page Not Found", http.StatusNotFound)
		return
	}

	fileData := make(map[string]any)
	fileData["file_id"] = id
	fileData["filename"] = fileInfo.FileName
	fileData["description"] = fileInfo.Description
	fileData["type"] = fileInfo.FileType
	fileData["size"] = fileInfo.Size
	fileData["uploaded"] = fileInfo.Created

	data := app.newTemplateData(fileData)

	err = response.Page(w, http.StatusOK, data, "pages/file_info.tmpl")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) ReadFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	workDir, _ := os.Getwd()
	f, err := os.Open(filepath.Join(workDir, "uploads", id+".pdf"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	w.Header().Set("Content-type", "application/pdf")

	if _, err := io.Copy(w, f); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}
