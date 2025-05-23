package controllers

import (
	"database/sql"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/learninNdi/gotoko/app/consts"
	"github.com/learninNdi/gotoko/app/models"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/shopspring/decimal"
	"github.com/unrolled/render"
)

type CheckoutRequest struct {
	Cart            *models.Cart
	ShippingFee     *ShippingFee
	ShippingAddress *ShippingAddress
}

type ShippingFee struct {
	Courier     string
	PackageName string
	Fee         float64
}

type ShippingAddress struct {
	FirstName  string
	LastName   string
	CityID     string
	ProvinceID string
	Address1   string
	Address2   string
	Phone      string
	Email      string
	PostCode   string
}

func (server *Server) Checkout(w http.ResponseWriter, r *http.Request) {
	var cart *models.Cart

	cartID := GetShoppingCartID(w, r)
	cart, _ = GetShoppingCart(server.DB, cartID)
	items, _ := cart.GetItems(server.DB, cartID)

	if len(items) == 0 {
		SetFlash(w, r, "error", "Belum ada barang yang dipilih")
		http.Redirect(w, r, "/carts", http.StatusSeeOther)

		return
	}

	if !IsLoggedIn(r) {
		SetFlash(w, r, "error", "Anda harus login terlebih dahulu")
		http.Redirect(w, r, "/login", http.StatusSeeOther)

		return
	}

	user := server.CurrentUser(w, r)

	// hitung ulang
	shippingCost, err := server.GetSelectedShippingCost(w, r, cart)

	if err != nil {
		SetFlash(w, r, "error", "Proses checkout gagal")
		http.Redirect(w, r, "/carts", http.StatusSeeOther)

		return
	}

	checkoutRequest := &CheckoutRequest{
		Cart: cart,
		ShippingFee: &ShippingFee{
			Courier:     r.FormValue("courier"),
			PackageName: r.FormValue("shipping_fee"),
			Fee:         shippingCost,
		},
		ShippingAddress: &ShippingAddress{
			FirstName:  r.FormValue("first_name"),
			LastName:   r.FormValue("last_name"),
			CityID:     r.FormValue("city_id"),
			ProvinceID: r.FormValue("province_id"),
			Address1:   r.FormValue("address1"),
			Address2:   r.FormValue("address2"),
			Phone:      r.FormValue("phone"),
			Email:      r.FormValue("email"),
			PostCode:   r.FormValue("post_code"),
		},
	}

	order, err := server.SaveOrder(user, checkoutRequest)

	if err != nil {
		SetFlash(w, r, "error", "Proses checkout gagal")
		http.Redirect(w, r, "/carts", http.StatusSeeOther)

		return
	}

	ClearShoppingCart(server.DB, cartID)

	SetFlash(w, r, "success", "Data order berhasil disimpan")
	http.Redirect(w, r, "/orders/"+order.ID, http.StatusSeeOther)
}

func (server *Server) GetSelectedShippingCost(w http.ResponseWriter, r *http.Request, cart *models.Cart) (float64, error) {
	origin := os.Getenv("API_ONGKIR_ORIGIN")
	destination := r.FormValue("city_id")
	courier := r.FormValue("courier")
	shippingFeeSelected := r.FormValue("shipping_fee")

	shippingFeeOptions, err := server.CalculateShippingFee(models.ShippingFeeParams{
		Origin:      origin,
		Destination: destination,
		Courier:     courier,
		Weight:      cart.TotalWeight,
	})

	if err != nil {
		return 0, errors.New("Failed shipping calculation")
	}

	var shippingCost float64
	for _, shippingFeeOption := range shippingFeeOptions {
		if shippingFeeOption.Service == shippingFeeSelected {
			shippingCost = float64(shippingFeeOption.Fee)
		}
	}

	return shippingCost, nil
}

func (server *Server) SaveOrder(user *models.User, r *CheckoutRequest) (*models.Order, error) {
	var orderItems []models.OrderItem

	orderID := uuid.New().String()
	paymentURL, err := server.createPaymentURL(user, r, orderID)

	if err != nil {
		return nil, err
	}

	if len(r.Cart.CartItems) > 0 {
		for _, cartItem := range r.Cart.CartItems {
			orderItems = append(orderItems, models.OrderItem{
				ProductID:       cartItem.ProductID,
				Qty:             cartItem.Qty,
				BasePrice:       cartItem.BasePrice,
				BaseTotal:       cartItem.BaseTotal,
				TaxAmount:       cartItem.TaxAmount,
				TaxPercent:      cartItem.TaxPercent,
				DiscountAmount:  cartItem.DiscountAmount,
				DiscountPercent: cartItem.DiscountPercent,
				SubTotal:        cartItem.SubTotal,
				Sku:             cartItem.Product.Sku,
				Name:            cartItem.Product.Name,
				Weight:          cartItem.Product.Weight,
			})
		}
	}

	orderCustomer := &models.OrderCustomer{
		UserID:     user.ID,
		FirstName:  r.ShippingAddress.FirstName,
		LastName:   r.ShippingAddress.LastName,
		CityID:     r.ShippingAddress.CityID,
		ProvinceID: r.ShippingAddress.ProvinceID,
		Address1:   r.ShippingAddress.Address1,
		Address2:   r.ShippingAddress.Address2,
		Phone:      r.ShippingAddress.Phone,
		Email:      r.ShippingAddress.Email,
		PostCode:   r.ShippingAddress.PostCode,
	}

	orderData := &models.Order{
		ID:                  orderID,
		UserID:              user.ID,
		OrderItems:          orderItems,
		OrderCustomer:       orderCustomer,
		Status:              0,
		OrderDate:           time.Now(),
		PaymentDue:          time.Now().AddDate(0, 0, 1),
		PaymentStatus:       consts.OrderPaymentStatusUnpaid,
		BaseTotalPrice:      r.Cart.BaseTotalPrice,
		TaxAmount:           r.Cart.TaxAmount,
		TaxPercent:          r.Cart.TaxPercent,
		DiscountAmount:      r.Cart.DiscountAmount,
		DiscountPercent:     r.Cart.DiscountPercent,
		ShippingCost:        decimal.NewFromFloat(r.ShippingFee.Fee),
		GrandTotal:          r.Cart.GrandTotal,
		ShippingCourier:     r.ShippingFee.Courier,
		ShippingServiceName: r.ShippingFee.PackageName,
		PaymentToken:        sql.NullString{String: paymentURL, Valid: true},
	}

	orderModel := models.Order{}
	order, err := orderModel.CreateOrder(server.DB, orderData)

	if err != nil {
		return nil, err
	}

	return order, nil
}

func (server *Server) ShowOrder(w http.ResponseWriter, r *http.Request) {
	render := render.New(render.Options{
		Layout:     "layout",
		Extensions: []string{".html", ".tmpl"},
	})

	vars := mux.Vars(r)

	if vars["id"] == "" {
		http.Redirect(w, r, "/products", http.StatusSeeOther)

		return
	}

	orderModel := models.Order{}
	order, err := orderModel.FindByID(server.DB, vars["id"])

	if err != nil {
		http.Redirect(w, r, "/products", http.StatusSeeOther)

		return
	}

	_ = render.HTML(w, http.StatusOK, "show_order", map[string]interface{}{
		"order":   order,
		"success": GetFlash(w, r, "success"),
		"user":    server.CurrentUser(w, r),
	})
}

func (server *Server) createPaymentURL(user *models.User, r *CheckoutRequest, orderID string) (string, error) {
	midtransServerKey := os.Getenv("API_MIDTRANS_KEY")

	midtrans.ServerKey = midtransServerKey

	var enabledPaymentTypes []snap.SnapPaymentType
	enabledPaymentTypes = append(enabledPaymentTypes, snap.AllSnapPaymentType...)

	snapRequest := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: r.Cart.GrandTotal.IntPart(),
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: user.FirstName,
			LName: user.LastName,
			Email: user.Email,
		},
		EnabledPayments: enabledPaymentTypes,
	}

	snapResponse, err := snap.CreateTransaction(snapRequest)

	if err != nil {
		return "", err
	}

	return snapResponse.RedirectURL, nil
}
