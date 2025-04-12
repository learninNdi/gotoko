package controllers

import (
	"net/http"

	"github.com/unrolled/render"
)

func (server *Server) Checkout(w http.ResponseWriter, r *http.Request) {
	render := render.New(render.Options{
		Layout:     "layout",
		Extensions: []string{".html", ".tmpl"},
	})

	if !IsLoggedIn(r) {
		SetFlash(w, r, "error", "Anda harus login terlebih dahulu")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	user := server.CurrentUser(w, r)

	_ = render.HTML(w, http.StatusOK, "checkout", map[string]interface{}{
		"user": user,
	})
}
