package main

import (
	"net/http"

	"github.com/maheshrc27/storemypdf/assets"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()
	r.NotFound(app.notFound)

	r.Use(app.logAccess)
	r.Use(app.recoverPanic)
	r.Use(app.securityHeaders)

	fileServer := http.FileServer(http.FS(assets.EmbeddedFiles))

	dir := http.Dir("./uploads")
	uploadsServer := http.FileServer(dir)

	r.Handle("/static/*", fileServer)
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", uploadsServer))

	r.Group(func(r chi.Router) {
		r.Use(app.AuthMiddleware)

		r.Get("/", app.home)
		r.Get("/docs", app.ApiDocs)
		r.Get("/f/{fileid}", app.FileInfo)
		r.Get("/f/{fileid}/open", app.ReadFile)
		r.Post("/f/{fileid}/download", app.DownloadFile)
		r.Get("/verify-email", app.VerifyEmail)

		r.Get("/u/files", app.ListFiles)
		r.Get("/u/dashboard", app.UserDashboard)
		r.Get("/u/account", app.UserAccount)
		r.Get("/u/subscription", app.Subscription)
		r.Get("/u/subscription/subscribe", app.Purchase)
		r.Get("/u/api-keys", app.ApiKeys)
		r.Post("/u/files/delete", app.DeleteFile)
		r.Post("/u/account/change-password", app.ChangePassword)
		r.Post("/u/account/delete", app.DeleteAccount)
		r.Post("/u/generate-api-key", app.GenerateApiKey)
		r.Post("/u/api-keys/delete", app.DeleteApiKey)

		r.Get("/signup", app.SignUp)
		r.Post("/signup", app.SignUp)
		r.Get("/signin", app.SignIn)
		r.Post("/signin", app.SignIn)
		r.Post("/signout", app.Logout)

		r.Post("/upload", app.UploadFile)
	})

	r.Group(func(r chi.Router) {
		r.Use(cors.Handler(cors.Options{
			// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
			AllowedOrigins: []string{"https://*", "http://*"},
			// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
			AllowedMethods:   []string{"POST"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}))

		r.Post("/paddle/webhook", app.PaddleWebhook)

		r.Route("/api", func(r chi.Router) {
			r.Use(app.ApiMiddleware)
			r.Post("/upload", app.UploadFileApi)
			// r.Get("/files/{file_id}", app.FileInfoApi)
		})
	})

	return r
}
