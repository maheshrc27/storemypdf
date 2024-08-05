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
	r.Use(app.AuthMiddleware)

	fileServer := http.FileServer(http.FS(assets.EmbeddedFiles))
	r.Handle("/static/*", fileServer)

	r.Get("/", app.home)
	r.Get("/api/docs", app.ApiDocs)
	r.Get("/f/{id}", app.FileInfo)
	r.Get("/f/{id}/open", app.ReadFile)

	r.Get("/signup", app.SignUp)
	r.Post("/signup", app.SignUp)
	r.Get("/signin", app.SignIn)
	r.Post("/signin", app.SignIn)
	r.Post("/signout", app.Logout)

	r.Post("/upload", app.UploadFile)

	r.Route("/api", func(router chi.Router) {
		r.Use(app.ApiMiddleware)
		router.Post("/upload", app.UploadFileApi)
		router.Get("/files/{file_id}", app.FileInfoApi)
	})

	return r
}
