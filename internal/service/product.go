package service

import (
	"code_base/internal/port/inbound/api"
	"code_base/internal/port/outbound/repository"
	"context"
	"errors"
)

type ProductService struct {
	repo repository.ProductRepository
}

func NewProductService(r repository.ProductRepository) ProductService {
	return ProductService{
		repo: r,
	}
}

func (p ProductService) AddProduct(ctx context.Context, cmd api.AddProductCommand) (api.AddProductResult, error) {
	if cmd.Product.Count <= 0 {
		return api.AddProductResult{}, errors.New("count must be greater than zero")
	}

	foundEntity, err := p.repo.FindByName(ctx, cmd.Product.Name)
	if err != nil {
		return api.AddProductResult{}, err
	}
	if foundEntity.Name != "" {
		return api.AddProductResult{}, errors.New("product already exists")
	}

	entityID, err := p.repo.Create(ctx, mapApiProductToEntityProduct(cmd.Product))
	if err != nil {
		return api.AddProductResult{}, err
	}

	return api.AddProductResult{ID: entityID}, nil
}

func (p ProductService) ListProducts(ctx context.Context, query api.ListProductsQuery) (api.ListProductsResult, error) {
	productsList, err := p.repo.GetAll(ctx)
	if err != nil {
		return api.ListProductsResult{}, err
	}

	return api.ListProductsResult{
		Products: mapEntityProductsToApiProducts(productsList),
	}, err
}
