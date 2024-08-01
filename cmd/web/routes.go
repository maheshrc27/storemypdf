package main

import (
	"net/http"

	"github.com/maheshrc27/storemypdf/assets"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()
	r.NotFound(app.notFound)

	r.Use(app.logAccess)
	r.Use(app.recoverPanic)
	r.Use(app.securityHeaders)

	fileServer := http.FileServer(http.FS(assets.EmbeddedFiles))
	r.Handle("/static/*", fileServer)

	r.Get("/", app.home)
	r.Get("/api/docs", app.ApiDocs)

	r.Get("/f/{id}", app.FileInfo)

	r.Get("/signup", app.SignUp)
	r.Post("/signup", app.SignUp)

	r.Get("/signin", app.SignIn)
	r.Post("/signin", app.SignIn)

	r.Post("/upload", app.UploadFile)

	return r
}
