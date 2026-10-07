package repository

import (
	"code_base/internal/domain/entity"
	"context"
	"errors"
)

type PostgresProductRepository struct {
}

func NewPostgresProductRepository() PostgresProductRepository {
	return PostgresProductRepository{}
}

func (p PostgresProductRepository) Create(context.Context, entity.Product) (string, error) {
	return "", errors.New("not implemented")
}

func (p PostgresProductRepository) FindByID(context.Context, string) (entity.Product, error) {
	return entity.Product{}, errors.New("not implemented")
}

func (p PostgresProductRepository) FindByName(context.Context, string) (entity.Product, error) {
	return entity.Product{}, errors.New("not implemented")
}

func (p PostgresProductRepository) GetAll(context.Context) ([]entity.Product, error) {
	return []entity.Product{}, errors.New("not implemented")
}
