package controllers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/learninNdi/gotoko/app/models"
	"github.com/unrolled/render"
)

func (server *Server) Login(w http.ResponseWriter, r *http.Request) {
	render := render.New(render.Options{
		Layout:     "layout",
		Extensions: []string{".html", ".tmpl"},
	})

	_ = render.HTML(w, http.StatusOK, "login", map[string]interface{}{
		"error":   GetFlash(w, r, "error"),
		"success": GetFlash(w, r, "success"),
	})
}

func (server *Server) DoLogin(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	userModel := models.User{}
	user, err := userModel.FindByEmail(server.DB, email)

	if err != nil {
		SetFlash(w, r, "error", "email invalid")
		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}

	if !ComparePassword(password, user.Password) {
		SetFlash(w, r, "error", "password invalid")
		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}

	session, _ := store.Get(r, sessionUser)
	session.Values["id"] = user.ID
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (server *Server) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, sessionUser)

	session.Values["id"] = nil
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (server *Server) Register(w http.ResponseWriter, r *http.Request) {
	render := render.New(render.Options{
		Layout:     "layout",
		Extensions: []string{".html", ".tmpl"},
	})

	_ = render.HTML(w, http.StatusOK, "register", map[string]interface{}{
		"error": GetFlash(w, r, "error"),
	})
}

func (server *Server) DoRegister(w http.ResponseWriter, r *http.Request) {
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if firstName == "" || lastName == "" || email == "" || password == "" {
		if firstName == "" {
			SetFlash(w, r, "error", "First name invalid")
		}

		if lastName == "" {
			SetFlash(w, r, "error", "Last name invalid")
		}

		if email == "" {
			SetFlash(w, r, "error", "Email invalid")
		}

		if password == "" {
			SetFlash(w, r, "error", "Password invalid")
		}

		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	userModel := models.User{}
	existUser, _ := userModel.FindByEmail(server.DB, email)

	if existUser != nil {
		SetFlash(w, r, "error", "Email already registered")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	// hashedPassword, _ := MakePassword(password)

	params := &models.User{
		ID:        uuid.New().String(),
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		// Password: hashedPassword,
		Password: password,
	}

	_, err := userModel.CreateUser(server.DB, params)
	// user, err := userModel.CreateUser(server.DB, params)

	if err != nil {
		SetFlash(w, r, "error", "Registration failed")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	// session, _ := store.Get(r, sessionUser)
	// session.Values["id"] = user.ID
	// session.Save(r, w)

	// http.Redirect(w, r, "/", http.StatusSeeOther)

	SetFlash(w, r, "success", "Registration success")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
