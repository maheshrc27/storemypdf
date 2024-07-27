package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/maheshrc27/storemypdf/internal/cookies"
	"github.com/maheshrc27/storemypdf/internal/request"
	"github.com/maheshrc27/storemypdf/internal/response"
	"golang.org/x/crypto/bcrypt"
)

func (app *application) SignUp(w http.ResponseWriter, r *http.Request) {

	type createAccountForm struct {
		Email           string `form:"email"`
		Password        string `form:"password"`
		ConfirmPassword string `form:"confirm_password"`
	}

	switch r.Method {
	case http.MethodGet:
		data := app.newTemplateData(r)

		err := response.Page(w, http.StatusOK, data, "pages/signup.tmpl")
		if err != nil {
			app.serverError(w, r, err)
		}

	case http.MethodPost:
		var form createAccountForm

		err := request.DecodePostForm(r, &form)
		if err != nil {
			app.badRequest(w, r, err)
			return
		}

		if form.Password != form.ConfirmPassword {
			http.Error(w, "Passwords do not match", http.StatusBadRequest)
			return
		}

		_, found, err := app.db.GetUserByEmail(form.Email)
		if found {
			fmt.Println(err)
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}

		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = app.db.InsertUser(form.Email, string(hashedPassword))
		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/signin", http.StatusFound)

	}
}

func (app *application) SignIn(w http.ResponseWriter, r *http.Request) {
	type Credentials struct {
		Email    string `form:"email"`
		Password string `form:"password"`
	}

	type Claims struct {
		Id string `json:"id"`
		jwt.StandardClaims
	}

	switch r.Method {
	case http.MethodGet:
		data := app.newTemplateData(r)
		err := response.Page(w, http.StatusOK, data, "pages/signin.tmpl")
		if err != nil {
			app.serverError(w, r, err)
		}

	case http.MethodPost:
		var form Credentials

		err := request.DecodePostForm(r, &form)
		if err != nil {
			app.badRequest(w, r, err)
			return
		}

		result, found, err := app.db.GetUserByEmail(form.Email)
		fmt.Println(result)
		if err != nil {
			log.Printf("Error fetching user by email: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !found {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(result.HashedPassword), []byte(form.Password)); err != nil {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		expirationTime := time.Now().Add(24 * time.Hour)
		claims := &Claims{
			Id: strconv.Itoa(result.ID),
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(app.config.cookie.secretKey))
		if err != nil {
			log.Printf("Error signing JWT: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		cookie := http.Cookie{
			Name:     "authentication",
			Value:    tokenString,
			Path:     "/",
			MaxAge:   int(time.Until(expirationTime).Seconds()),
			HttpOnly: true,
			Secure:   true,
		}

		cookies.Write(w, cookie)

		http.Redirect(w, r, "/", http.StatusFound)
	}
}
