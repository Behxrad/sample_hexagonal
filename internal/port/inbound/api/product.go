package api

import (
	"context"
)

type AddProductCommand struct {
	Product Product
}

type AddProductResult struct {
	ID string
}

type ListProductsQuery struct{}

type ListProductsResult struct {
	Products []Product
}

type Product struct {
	Count uint16
	Price uint32
	ID    string
	Name  string
}

type ProductService interface {
	AddProduct(context.Context, AddProductCommand) (AddProductResult, error)
	ListProducts(context.Context, ListProductsQuery) (ListProductsResult, error)
}
