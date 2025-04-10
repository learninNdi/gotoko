package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/learninNdi/gotoko/app/models"
	"github.com/shopspring/decimal"
)

func (server *Server) GetProvinces() ([]models.Province, error) {
	response, err := http.Get(os.Getenv("API_ONGKIR_BASE_URL") + "province?key=" + os.Getenv("API_ONGKIR_KEY"))

	if err != nil {
		log.Fatal(err)

		return nil, err
	}

	provinceResponse := models.ProvinceResponse{}
	body, readErr := io.ReadAll(response.Body)

	if readErr != nil {
		return nil, readErr
	}

	jsonErr := json.Unmarshal(body, &provinceResponse)

	if jsonErr != nil {
		return nil, jsonErr
	}

	return provinceResponse.ProvinceData.Results, nil
}

func (server *Server) GetCitiesByProvinceID(provinceID string) ([]models.City, error) {
	response, err := http.Get(os.Getenv("API_ONGKIR_BASE_URL") + "city?key=" + os.Getenv("API_ONGKIR_KEY") + "&province=" + provinceID)

	if err != nil {
		log.Fatal(err)

		return nil, err
	}

	cityResponse := models.CityResponse{}

	body, readErr := io.ReadAll(response.Body)

	if readErr != nil {
		return nil, err
	}

	jsonErr := json.Unmarshal(body, &cityResponse)

	if jsonErr != nil {
		return nil, err
	}

	return cityResponse.CityData.Results, nil
}

func (server *Server) GetCitiesByProvince(w http.ResponseWriter, r *http.Request) {
	provinceID := r.URL.Query().Get("province_id")

	cities, err := server.GetCitiesByProvinceID(provinceID)

	if err != nil {
		log.Fatal(err)
	}

	res := Result{Code: 20, Data: cities, Message: "Success"}
	result, err := json.Marshal(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

func (server *Server) CalculateShipping(w http.ResponseWriter, r *http.Request) {
	origin := os.Getenv("API_ONGKIR_ORIGIN")
	destination := r.FormValue("city_id")
	courier := r.FormValue("courier")

	if destination == "" {
		http.Error(w, "invalid destination", http.StatusInternalServerError)
	}

	cartID := GetShoppingCartID(w, r)
	cart, _ := GetShoppingCart(server.DB, cartID)

	shippingFeeOptions, err := server.CalculateShippingFee(models.ShippingFeeParams{
		Origin:      origin,
		Destination: destination,
		Courier:     courier,
		Weight:      cart.TotalWeight,
	})

	if err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	res := Result{Code: 200, Data: shippingFeeOptions, Message: "Success"}
	result, _ := json.Marshal(res)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

func (server *Server) CalculateShippingFee(shippingParams models.ShippingFeeParams) ([]models.ShippingFeeOption, error) {
	if shippingParams.Origin == "" || shippingParams.Destination == "" || shippingParams.Weight <= 0 || shippingParams.Courier == "" {
		return nil, errors.New("invalid params")
	}

	params := url.Values{}
	params.Add("key", os.Getenv("API_ONGKIR_KEY"))
	params.Add("origin", shippingParams.Origin)
	params.Add("destination", shippingParams.Destination)
	params.Add("weight", strconv.Itoa(shippingParams.Weight))
	params.Add("courier", shippingParams.Courier)

	response, err := http.PostForm(os.Getenv("API_ONGKIR_BASE_URL")+"cost", params)

	if err != nil {
		return nil, err
	}

	ongkirResponse := models.OngkirResponse{}
	body, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	jsonError := json.Unmarshal(body, &ongkirResponse)

	if jsonError != nil {
		return nil, jsonError
	}

	var shippingFeeOptions []models.ShippingFeeOption

	for _, result := range ongkirResponse.OngkirData.Results {
		for _, cost := range result.Costs {
			shippingFeeOptions = append(shippingFeeOptions, models.ShippingFeeOption{
				Service: cost.Service,
				Fee:     cost.Cost[0].Value,
			})
		}
	}

	return shippingFeeOptions, nil
}

func (server *Server) ApplyShipping(w http.ResponseWriter, r *http.Request) {
	origin := os.Getenv("API_ONGKIR_ORIGIN")
	destination := r.FormValue("city_id")
	courier := r.FormValue("courier")
	shippingPackage := r.FormValue("shipping_package")

	cartID := GetShoppingCartID(w, r)
	cart, _ := GetShoppingCart(server.DB, cartID)

	if destination == "" {
		http.Error(w, "invalid destination", http.StatusInternalServerError)

		return
	}

	shippingFeeOptions, err := server.CalculateShippingFee(models.ShippingFeeParams{
		Origin:      origin,
		Destination: destination,
		Weight:      cart.TotalWeight,
		Courier:     courier,
	})

	if err != nil {
		http.Error(w, "invalid shipping calculation", http.StatusInternalServerError)

		return
	}

	var selectedShipping models.ShippingFeeOption

	for _, shippingOption := range shippingFeeOptions {
		if shippingOption.Service == shippingPackage {
			selectedShipping = shippingOption

			break
		}
	}

	cartGrandTotal, _ := cart.GrandTotal.Float64()
	shippingFee := float64(selectedShipping.Fee)
	grandTotal := cartGrandTotal + shippingFee

	applyShippingResponse := models.ApplyShippingResponse{
		TotalOrder:  cart.GrandTotal,
		ShippingFee: decimal.NewFromInt(selectedShipping.Fee),
		GrandTotal:  decimal.NewFromFloat(grandTotal),
		TotalWeight: decimal.NewFromInt(int64(cart.TotalWeight)),
	}

	res := Result{Code: 200, Data: applyShippingResponse, Message: "Success"}
	result, _ := json.Marshal(res)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}
