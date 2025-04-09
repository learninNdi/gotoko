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

}

func GetShoppingCartID(w http.ResponseWriter, r *http.Request) string {
	session, _ := store.Get(r, sessionShoppingCart)

	fmt.Println(session)

	if session.Values["cart-id"] == nil {

		session.Values["cart-id"] = uuid.New().String()
		_ = session.Save(r, w)
	} else {

	}

	return fmt.Sprintf("%v", session.Values["cart-id"])
}

func GetShoppingCart(db *gorm.DB, cartID string) (*models.Cart, error) {
	var cart models.Cart

	existCart, err := cart.GetCart(db, cartID)

	if err != nil {
		existCart, _ = cart.CreateCart(db, cartID)
	}

	existCart.CalculateCart(db, cartID)

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

	cartID := GetShoppingCartID(w, r)
	cart, _ := GetShoppingCart(server.DB, cartID)
	_, err = cart.AddItem(server.DB, models.CartItem{
		ProductID: productID,
		Qty:       qty,
	})

	if err != nil {
		http.Redirect(w, r, "/products/"+product.Slug, http.StatusSeeOther)
	}

	http.Redirect(w, r, "/carts", http.StatusSeeOther)
}
