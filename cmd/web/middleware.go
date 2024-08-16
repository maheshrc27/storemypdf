package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/maheshrc27/storemypdf/internal/cookies"
	"github.com/maheshrc27/storemypdf/internal/response"
	"github.com/maheshrc27/storemypdf/internal/tokens"

	"github.com/tomasen/realip"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				app.serverError(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")

		next.ServeHTTP(w, r)
	})
}

func (app *application) logAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mw := response.NewMetricsResponseWriter(w)
		next.ServeHTTP(mw, r)

		var (
			ip     = realip.FromRequest(r)
			method = r.Method
			url    = r.URL.String()
			proto  = r.Proto
		)

		userAttrs := slog.Group("user", "ip", ip)
		requestAttrs := slog.Group("request", "method", method, "url", url, "proto", proto)
		responseAttrs := slog.Group("repsonse", "status", mw.StatusCode, "size", mw.BytesCount)

		app.logger.Info("access", userAttrs, requestAttrs, responseAttrs)
	})
}

func (app *application) ApiMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.URL.Query().Get("key")

		if apiKey == "" {
			http.Error(w, "API key is required", http.StatusUnauthorized)
			return
		}

		userID, found, err := app.db.GetUserIDByKey(apiKey)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !found {
			http.Error(w, "Unauthorized Api Key", http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-User-ID", userID)

		next.ServeHTTP(w, r)
	})
}

func (app *application) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := cookies.Read(r, "authentication")

		if strings.HasPrefix(r.URL.Path, "/u/") && cookie == "" {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		if cookie != "" {
			authorized, err := tokens.IsAuthorized(cookie, app.config.cookie.secretKey)
			if err != nil || !authorized {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := tokens.ExtractIDFromToken(cookie, app.config.cookie.secretKey)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusBadRequest)
				return
			}

			r.Header.Set("X-Logged-IN", "true")
			r.Header.Set("X-User-ID", userID)
		}

		next.ServeHTTP(w, r)
	})
}
