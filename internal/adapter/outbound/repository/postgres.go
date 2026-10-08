package repository

import (
	"code_base/internal/domain/entity"
	"code_base/internal/port/outbound/repository"
	"context"
	"errors"
)

type postgresProductRepository struct {
}

func NewPostgresProductRepository() repository.ProductRepository {
	return postgresProductRepository{}
}

func (p postgresProductRepository) Create(context.Context, entity.Product) (string, error) {
	return "", errors.New("not implemented")
}

func (p postgresProductRepository) FindByID(context.Context, string) (entity.Product, error) {
	return entity.Product{}, errors.New("not implemented")
}

func (p postgresProductRepository) FindByName(context.Context, string) (entity.Product, error) {
	return entity.Product{}, errors.New("not implemented")
}

func (p postgresProductRepository) GetAll(context.Context) ([]entity.Product, error) {
	return []entity.Product{}, errors.New("not implemented")
}
