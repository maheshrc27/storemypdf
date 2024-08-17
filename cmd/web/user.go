package main

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/maheshrc27/storemypdf/internal/cookies"
	"github.com/maheshrc27/storemypdf/internal/request"
	"github.com/maheshrc27/storemypdf/internal/response"
	"github.com/maheshrc27/storemypdf/internal/tokens"
	"github.com/maheshrc27/storemypdf/templates/pages"
	"golang.org/x/crypto/bcrypt"
)

func (app *application) SignUp(w http.ResponseWriter, r *http.Request) {

	isLoggedIn := GetAuthStatus(r)
	if isLoggedIn {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	type createAccountForm struct {
		Email           string `form:"email"`
		Password        string `form:"password"`
		ConfirmPassword string `form:"confirm_password"`
	}

	switch r.Method {
	case http.MethodGet:
		signup := pages.SignUp("Create a Account")
		signup.Render(context.Background(), w)

	case http.MethodPost:
		var form createAccountForm

		err := request.DecodePostForm(r, &form)
		if err != nil {
			app.badRequest(w, r, err)
			return
		}

		if form.Password != form.ConfirmPassword {
			response.HTML(w, "Passowrd and Confirm Passowrd is not equal")
			return
		}

		_, found, err := app.db.GetUserByEmail(form.Email)
		if found {
			response.HTML(w, "User already exists")
			return
		}

		if err != nil {
			response.HTML(w, err.Error())
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
		if err != nil {
			response.HTML(w, "Internal server error")
			return
		}

		_, err = app.db.InsertUser(form.Email, string(hashedPassword))
		if err != nil {
			response.HTML(w, "Couldn't create user")
			return
		}

		w.Header().Set("HX-Redirect", "/signin")
	}
}

func (app *application) SignIn(w http.ResponseWriter, r *http.Request) {

	isLoggedIn := GetAuthStatus(r)
	if isLoggedIn {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	type Credentials struct {
		Email    string `form:"email"`
		Password string `form:"password"`
	}

	switch r.Method {
	case http.MethodGet:
		signin := pages.SignIn("Sign In to your Account")
		signin.Render(context.Background(), w)

	case http.MethodPost:
		var form Credentials

		err := request.DecodePostForm(r, &form)
		if err != nil {
			app.badRequest(w, r, err)
			return
		}

		result, found, err := app.db.GetUserByEmail(form.Email)
		if !found {
			response.HTML(w, "Invalid email or password")
			return
		}

		if err != nil {
			response.HTML(w, "Invalid email or password")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(result.HashedPassword), []byte(form.Password)); err != nil {
			response.HTML(w, "Invalid email or password")
			return
		}

		tokenString, err := tokens.GenerateJWT(result.ID.String(), app.config.cookie.secretKey)
		if err != nil {
			response.HTML(w, "Internal Server Error")
			return
		}

		cookie := http.Cookie{
			Name:     "authentication",
			Value:    tokenString,
			Path:     "/",
			MaxAge:   int(time.Until(time.Now().Add(24 * time.Hour)).Seconds()),
			HttpOnly: true,
			Secure:   false,
		}

		cookies.Write(w, cookie)

		w.Header().Set("HX-Redirect", "/u/dashboard")
	}
}

func (app *application) Logout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "authentication",
		Value:    "",
		Path:     "/",
		Expires:  time.Now(),
		HttpOnly: true,
		Secure:   true,
	}
	cookies.Write(w, cookie)

	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}

func (app *application) ChangePassword(w http.ResponseWriter, r *http.Request) {

	userID, err := ParseUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	type changePasswordForm struct {
		OldPassword        string `form:"old_password"`
		NewPassword        string `form:"new_password"`
		ConfirmNewPassword string `form:"confirm_new_password"`
	}

	var form changePasswordForm

	err = request.DecodePostForm(r, &form)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	user, found, err := app.db.GetUser(userID)
	if !found {
		http.Error(w, "User Not Found", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(form.OldPassword)); err != nil {
		http.Error(w, "Wrong Passowrd", http.StatusUnauthorized)
		return
	}

	if form.NewPassword != form.ConfirmNewPassword {
		response.HTML(w, "Password and Confirm Passowrd do not match")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.HTML(w, "Wrong password")
		return
	}

	err = app.db.UpdateUserHashedPassword(userID, string(hashedPassword))
	if err != nil {
		response.HTML(w, "Internal Server Error")
		return
	}

	w.Header().Set("HX-Refresh", "true")
}

func (app *application) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	uid := r.Header.Get("X-User-ID")
	userID, err := uuid.Parse(uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	type deleteAccountForm struct {
		Passowrd string `form:"password"`
	}

	var form deleteAccountForm

	err = request.DecodePostForm(r, &form)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	user, found, err := app.db.GetUser(userID)
	if !found {
		http.Error(w, "User Not Found", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(form.Passowrd)); err != nil {
		response.HTML(w, "Wrong Password")
		return
	}

	err = app.db.DeleteUser(userID)
	if err != nil {
		response.HTML(w, "Couldn't delete user")
		return
	}

	app.Logout(w, r)

}
