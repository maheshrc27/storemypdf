package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/maheshrc27/storemypdf/templates/pages"
)

// front pages

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	home := pages.Home("store files", GetAuthStatus(r))
	home.Render(context.Background(), w)
}

func (app *application) ApiDocs(w http.ResponseWriter, r *http.Request) {
	page := pages.Docs("Account Settings", GetAuthStatus(r))
	page.Render(context.Background(), w)
}

func (app *application) FileInfo(w http.ResponseWriter, r *http.Request) {
	fileid := chi.URLParam(r, "fileid")

	fileInfo, found, err := app.db.GetFile(fileid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !found {
		http.Error(w, "Page Not Found", http.StatusNotFound)
		return
	}

	fileSize := float64(fileInfo.Size) / (1024 * 1024)

	page := pages.FIleInfo(fileInfo.FileName, GetAuthStatus(r), fileid, fileInfo.FileName, fileInfo.Description,
		fileInfo.FileType, fmt.Sprintf("%.2f MB", fileSize), fileInfo.Created.Format("January 2, 2006"))
	page.Render(context.Background(), w)
}

func (app *application) ReadFile(w http.ResponseWriter, r *http.Request) {
	fileid := chi.URLParam(r, "fileid")
	workDir, _ := os.Getwd()
	f, err := os.Open(filepath.Join(workDir, "uploads", fileid+".pdf"))
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

//dashboard pages

func (app *application) ListFiles(w http.ResponseWriter, r *http.Request) {

	userId, err := ParseUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	files, _, err := app.db.GetFilesByUserID(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	page := pages.ListFiles("uploads lists", files)
	page.Render(context.Background(), w)

}

func (app *application) Purchase(w http.ResponseWriter, r *http.Request) {

	userId := r.Header.Get("X-User-ID")

	page := pages.Payment(userId)
	page.Render(context.Background(), w)
}

func (app *application) UserDashboard(w http.ResponseWriter, r *http.Request) {
	page := pages.Dashboard("User Dashboard")
	page.Render(context.Background(), w)
}

func (app *application) UserAccount(w http.ResponseWriter, r *http.Request) {
	page := pages.Account("Account Settings")
	page.Render(context.Background(), w)
}

func (app *application) ApiKeys(w http.ResponseWriter, r *http.Request) {
	userId, err := ParseUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	keys, _, err := app.db.GetKeysByUserID(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	page := pages.ListKeys("API Keys", keys)
	page.Render(context.Background(), w)
}

func (app *application) Subscription(w http.ResponseWriter, r *http.Request) {
	userId, err := ParseUserID(r)
	if err != nil {
		handleError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	subscription, premium, err := CheckPremium(userId, app.db)
	if err != nil {
		handleError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	status := subscription.Status
	nextbilldate := subscription.NextBillDate.Format("January 2, 2006")

	page := pages.Subscription("Subscription Management", premium, status, nextbilldate)
	page.Render(context.Background(), w)
}
