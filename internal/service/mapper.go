package service

import (
	"code_base/internal/domain/entity"
	"code_base/internal/port/inbound/api"
)

func mapEntityProductsToApiProducts(products []entity.Product) []api.Product {
	productResult := make([]api.Product, 0)

	for _, product := range products {
		productResult = append(productResult, mapEntityProductToApiProduct(product))
	}

	return productResult
}

func mapEntityProductToApiProduct(product entity.Product) api.Product {
	return api.Product{
		Count: product.Count,
		Price: product.Price,
		ID:    product.ID,
		Name:  product.Name,
	}
}

func mapApiProductToEntityProduct(product api.Product) entity.Product {
	return entity.Product{
		Count: product.Count,
		Price: product.Price,
		ID:    product.ID,
		Name:  product.Name,
	}
}
