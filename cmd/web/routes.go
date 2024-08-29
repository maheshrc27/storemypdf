package main

import (
	"net/http"
	"time"

	"github.com/maheshrc27/storemypdf/assets"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()
	r.NotFound(app.notFound)

	r.Use(app.logAccess)
	r.Use(app.recoverPanic)
	r.Use(app.securityHeaders)

	r.Use(httprate.LimitByIP(50, 1*time.Minute))

	fileServer := http.FileServer(http.FS(assets.EmbeddedFiles))

	r.Handle("/static/*", fileServer)

	r.Group(func(r chi.Router) {
		r.Use(app.AuthMiddleware)

		r.Get("/", app.home)
		r.Get("/docs", app.ApiDocs)
		r.Get("/f/{fileid}", app.FileInfo)
		r.Get("/f/{fileid}/open", app.ReadFile)
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
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"POST"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300,
		}))

		r.Post("/paddle/webhook", app.PaddleWebhook)

		r.Route("/api", func(r chi.Router) {
			r.Use(app.ApiMiddleware)
			r.Post("/upload", app.UploadFileApi)
		})
	})

	return r
}
