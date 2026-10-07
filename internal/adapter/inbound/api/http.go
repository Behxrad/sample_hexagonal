package api

import (
	"code_base/internal/port/inbound/api"
	"encoding/json"
	"net/http"
)

type HttpServer struct {
	mux *http.ServeMux
}

func NewHttpServer(s api.ProductService) HttpServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/create-product", CreateProduct(s))
	mux.HandleFunc("/products", GetProducts(s))

	return HttpServer{
		mux: mux,
	}
}

func (h HttpServer) Start() {
	err := http.ListenAndServe(":8080", h.mux)
	if err != nil {
		return
	}
}

func CreateProduct(service api.ProductService) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		product := api.Product{}
		err := json.NewDecoder(r.Body).Decode(&product)
		if err != nil {
			writeError(w, err)
			return
		}
		addProduct, err := service.AddProduct(nil, api.AddProductCommand{Product: product})
		if err != nil {
			writeError(w, err)
			return
		}
		err = json.NewEncoder(w).Encode(addProduct)
		if err != nil {
			writeError(w, err)
			return
		}
	}
}

func GetProducts(service api.ProductService) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		products, err := service.ListProducts(r.Context(), api.ListProductsQuery{})
		if err != nil {
			writeError(w, err)
			return
		}
		err = json.NewEncoder(w).Encode(products)
		if err != nil {
			writeError(w, err)
			return
		}
	}
}

func writeError(w http.ResponseWriter, err error) {
	json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
}
