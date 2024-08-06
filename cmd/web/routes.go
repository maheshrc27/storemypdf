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

	r.Route("/api", func(router chi.Router) {
		r.Use(app.AuthMiddleware)

		router.Get("/", app.home)
		router.Get("/api/docs", app.ApiDocs)
		router.Get("/f/{id}", app.FileInfo)
		router.Get("/f/{id}/open", app.ReadFile)

		router.Get("/signup", app.SignUp)
		router.Post("/signup", app.SignUp)
		router.Get("/signin", app.SignIn)
		router.Post("/signin", app.SignIn)
		router.Post("/signout", app.Logout)

		router.Post("/upload", app.UploadFile)
	})

	r.Route("/api", func(router chi.Router) {
		router.Use(app.ApiMiddleware)
		router.Post("/upload", app.UploadFileApi)
		router.Get("/files/{file_id}", app.FileInfoApi)
	})

	return r
}
