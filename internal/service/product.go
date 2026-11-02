package service

import (
	"context"

	"github.com/Behxrad/sample_hexagonal/internal/errcodes"
	"github.com/Behxrad/sample_hexagonal/internal/port/inbound/api"
	"github.com/Behxrad/sample_hexagonal/internal/port/outbound/repository"

	"192.168.100.6/coreinno/common.git/errs"
)

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(r repository.ProductRepository) api.ProductService {
	return productService{
		repo: r,
	}
}

func (p productService) AddProduct(ctx context.Context, cmd api.AddProductCommand) (api.AddProductResult, error) {
	if cmd.Product.Count <= 0 {
		return api.AddProductResult{}, errs.NewError(errcodes.ErrInvalidProductCount)
	}

	foundEntity, err := p.repo.FindByName(ctx, cmd.Product.Name)
	if err != nil {
		return api.AddProductResult{}, err
	}
	if foundEntity.Name != "" {
		return api.AddProductResult{}, errs.NewError(errcodes.ErrProductExists)
	}

	entityID, err := p.repo.Create(ctx, mapApiProductToEntityProduct(cmd.Product))
	if err != nil {
		return api.AddProductResult{}, err
	}

	return api.AddProductResult{ID: entityID}, nil
}

func (p productService) ListProducts(ctx context.Context, query api.ListProductsQuery) (api.ListProductsResult, error) {
	productsList, err := p.repo.GetAll(ctx)
	if err != nil {
		return api.ListProductsResult{}, err
	}

	return api.ListProductsResult{
		Products: mapEntityProductsToApiProducts(productsList),
	}, err
}
