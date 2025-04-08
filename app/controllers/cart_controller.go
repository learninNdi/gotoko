package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/learninNdi/gotoko/app/models"
	"gorm.io/gorm"
)

func (server *Server) GetCart(w http.ResponseWriter, r *http.Request) {
	var cart *models.Cart

	cartID := GetShoppingCartID(w, r)

	fmt.Println("cart id !==== ", cartID)

	cart, _ = GetShoppingCart(server.DB, cartID)

	fmt.Println("cart id ===== ", cart.ID)
}

func GetShoppingCartID(w http.ResponseWriter, r *http.Request) string {
	session, _ := store.Get(r, "shopping-cart-session")

	fmt.Println(session)

	if session.Values["cart-id"] == nil {

		session.Values["cart-id"] = uuid.New().String()
		_ = session.Save(r, w)
	} else {

	}

	fmt.Println(session)
	fmt.Println(session.Values["cart-id"])

	return fmt.Sprintf("%v", session.Values["cart-id"])
}

func GetShoppingCart(db *gorm.DB, cartID string) (*models.Cart, error) {
	var cart models.Cart

	existCart, err := cart.GetCart(db, cartID)

	if err != nil {
		// fmt.Println("kosong")
		existCart, _ = cart.CreateCart(db, cartID)
	} else {
		// fmt.Println("ga kosong")
	}
	// fmt.Println(existCart)
	return existCart, nil
}

func (server *Server) AddItemToCart(w http.ResponseWriter, r *http.Request) {
	productID := r.FormValue("product_id")
	qty, _ := strconv.Atoi(r.FormValue("qty"))

	productModel := models.Product{}
	product, err := productModel.GetProductByID(server.DB, productID)

	if err != nil {
		http.Redirect(w, r, "/products/"+product.Slug, http.StatusSeeOther)
	}

	if qty > product.Stock {
		http.Redirect(w, r, "/products/"+product.Slug, http.StatusSeeOther)
	}

	http.Redirect(w, r, "/carts", http.StatusSeeOther)
}
