package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Behxrad/sample_hexagonal/configs"
	"github.com/Behxrad/sample_hexagonal/internal/port/inbound/api"
)

type HttpServer struct {
	mux *http.ServeMux
}

func NewHttpServer(s api.ProductService) HttpServer {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /create-product", CreateProduct(s))
	mux.HandleFunc("GET /products", GetProducts(s))

	return HttpServer{
		mux: mux,
	}
}

func (h HttpServer) Serve(ctx context.Context) error {
	go func() {
		server := &http.Server{
			Addr: fmt.Sprintf("%s:%d", configs.ApplicationConfigs.HTTPHost,
				configs.ApplicationConfigs.HTTPPort),
			Handler:           h.mux,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		}
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()
	return nil
}

func (h HttpServer) Close(ctx context.Context) error {
	log.Println("http server closed")
	return nil
}

func CreateProduct(service api.ProductService) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		product := api.Product{}
		err := json.NewDecoder(r.Body).Decode(&product)
		if err != nil {
			writeError(w, err)
			return
		}
		addProduct, err := service.AddProduct(r.Context(), api.AddProductCommand{Product: product})
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
		w.Header().Set("Content-Type", "application/json")
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
	w.Header().Set("Content-Type", "application/json")
	inError := json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
	if inError != nil {
		log.Println(inError)
		return
	}
}
