package controllers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/learninNdi/gotoko/app/models"
	"github.com/unrolled/render"
	"gorm.io/gorm"
)

func (server *Server) GetCart(w http.ResponseWriter, r *http.Request) {
	render := render.New(render.Options{
		Layout:     "layout",
		Extensions: []string{".html", ".tmpl"},
	})

	var cart *models.Cart

	cartID := GetShoppingCartID(w, r)
	cart, _ = GetShoppingCart(server.DB, cartID)
	items, _ := cart.GetItems(server.DB, cartID)

	provinces, err := server.GetProvinces()

	if err != nil {
		log.Fatal(err)
	}

	_ = render.HTML(w, http.StatusOK, "cart", map[string]interface{}{
		"cart":      cart,
		"items":     items,
		"provinces": provinces,
	})
}

func GetShoppingCartID(w http.ResponseWriter, r *http.Request) string {
	session, _ := store.Get(r, sessionShoppingCart)

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

	updatedCart, _ := cart.GetCart(db, cartID)

	totalWeight := 0
	productModel := models.Product{}

	for _, cartItem := range updatedCart.CartItems {
		product, _ := productModel.GetProductByID(db, cartItem.ProductID)

		productWeight, _ := product.Weight.Float64()
		itemWeight := math.Ceil(productWeight * float64(cartItem.Qty))

		totalWeight += int(itemWeight)
	}

	updatedCart.TotalWeight = totalWeight

	return updatedCart, nil
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

func (server *Server) UpdateCart(w http.ResponseWriter, r *http.Request) {
	cartID := GetShoppingCartID(w, r)
	cart, _ := GetShoppingCart(server.DB, cartID)

	for _, item := range cart.CartItems {
		qty, _ := strconv.Atoi(r.FormValue(item.ID))

		_, err := cart.UpdateItemQty(server.DB, item.ID, qty)

		if err != nil {
			http.Redirect(w, r, "/carts", http.StatusSeeOther)
		}
	}

	http.Redirect(w, r, "/carts", http.StatusSeeOther)
}

func (server *Server) RemoveItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if vars["id"] == "" {
		http.Redirect(w, r, "/carts", http.StatusSeeOther)
	}

	cartID := GetShoppingCartID(w, r)
	cart, _ := GetShoppingCart(server.DB, cartID)

	err := cart.RemoveItemByID(server.DB, vars["id"])

	if err != nil {
		http.Redirect(w, r, "/carts", http.StatusSeeOther)
	}

	http.Redirect(w, r, "/carts", http.StatusSeeOther)
}
