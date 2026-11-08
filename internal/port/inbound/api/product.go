package api

import (
	"context"

	inbound_headers "192.168.100.6/coreinno/proto_common.git/model/common/v1"
)

type AddProductCommand struct {
	MetaData inbound_headers.RequestMetadata
	Product  Product
}

type AddProductResult struct {
	MetaData inbound_headers.ResponseMetadata
	ID       string
}

type ListProductsQuery struct {
	MetaData inbound_headers.RequestMetadata
}

type ListProductsResult struct {
	MetaData inbound_headers.ResponseMetadata
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
