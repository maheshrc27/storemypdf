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

	r.Group(func(r chi.Router) {
		r.Use(app.AuthMiddleware)

		r.Get("/", app.home)
		r.Get("/docs", app.ApiDocs)
		r.Get("/f/{id}", app.FileInfo)
		r.Get("/f/{id}/open", app.ReadFile)
		r.Post("/f/{id}/download", app.FileDownload)

		r.Get("/u/uploads", app.ListFiles)

		r.Get("/signup", app.SignUp)
		r.Post("/signup", app.SignUp)
		r.Get("/signin", app.SignIn)
		r.Post("/signin", app.SignIn)
		r.Post("/signout", app.Logout)

		r.Post("/upload", app.UploadFile)
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(app.ApiMiddleware)
		r.Post("/upload", app.UploadFileApi)
		r.Get("/files/{file_id}", app.FileInfoApi)
	})

	return r
}
