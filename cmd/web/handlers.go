package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/maheshrc27/storemypdf/internal/response"
	"github.com/maheshrc27/storemypdf/templates/pages"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {

	isLoggedIn := r.Header.Get("X-Logged-IN")
	userId := r.Header.Get("X-User-ID")
	fmt.Println(isLoggedIn)
	fmt.Println(userId)

	home := pages.Home("store files", false)
	home.Render(context.Background(), w)
}

func (app *application) ApiDocs(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	err := response.Page(w, http.StatusOK, data, "pages/docs.tmpl")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) FileInfo(w http.ResponseWriter, r *http.Request) {
	isLoggedIn := r.Header.Get("X-Logged-IN")
	userId := r.Header.Get("X-User-ID")
	fmt.Println(isLoggedIn)
	fmt.Println(userId)
	id := chi.URLParam(r, "id")

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
	if isLoggedIn == "true" {
		fileData["loggedIn"] = true
	} else {
		fileData["loggedIn"] = false
	}
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
