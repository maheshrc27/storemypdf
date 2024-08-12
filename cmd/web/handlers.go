package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/maheshrc27/storemypdf/internal/response"
	"github.com/maheshrc27/storemypdf/templates/pages"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {

	isLoggedIn := r.Header.Get("X-Logged-IN")
	// userId := r.Header.Get("X-User-ID")

	loggedIn := false

	if isLoggedIn == "true" {
		loggedIn = true
	}

	home := pages.Home("store files", loggedIn)
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
	loggedIn := false

	if isLoggedIn == "true" {
		loggedIn = true
	}

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

	fileSize := float64(fileInfo.Size) / (1024 * 1024)

	home := pages.FIleInfo("store files", loggedIn, id, fileInfo.FileName, fileInfo.Description,
		fileInfo.FileType, fmt.Sprintf("%.2f MB", fileSize), fileInfo.Created.Format("January 2, 2006"))
	home.Render(context.Background(), w)
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

func (app *application) ListFiles(w http.ResponseWriter, r *http.Request) {
	uid := r.Header.Get("X-User-ID")
	userId, err := uuid.Parse(uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	files, _, err := app.db.GetFilesByUserID(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lists := pages.ListFiles("uploads lists", false, files)
	lists.Render(context.Background(), w)
}

func (app *application) FileDownload(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	workDir, _ := os.Getwd()

	fileInfo, found, err := app.db.GetFile(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !found {
		http.Error(w, "Page Not Found", http.StatusNotFound)
		return
	}

	filepath := filepath.Join(workDir, "uploads", id+".pdf")

	w.Header().Set("Content-Disposition", "attachment; filename="+fileInfo.FileName)
	w.Header().Set("Content-Type", fileInfo.FileType)
	http.ServeFile(w, r, filepath)
}

func (app *application) Purchase(w http.ResponseWriter, r *http.Request) {
	isLoggedIn := r.Header.Get("X-Logged-IN")
	userID := r.Header.Get("X-User-ID")
	loggedIn := false

	if isLoggedIn == "true" {
		loggedIn = true
	} else {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
	}

	fmt.Println(loggedIn)

	page := pages.Payment(userID)
	page.Render(context.Background(), w)
}

func (app *application) UserDashboard(w http.ResponseWriter, r *http.Request) {
	page := pages.Dashboard("User Dashboard")
	page.Render(context.Background(), w)
}
