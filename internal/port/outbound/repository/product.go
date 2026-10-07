package repository

import (
	"code_base/internal/domain/entity"
	"context"
)

type ProductRepository interface {
	Create(context.Context, entity.Product) (string, error)
	FindByID(context.Context, string) (entity.Product, error)
	FindByName(context.Context, string) (entity.Product, error)
	GetAll(context.Context) ([]entity.Product, error)
}
