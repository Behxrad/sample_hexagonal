package repository

import (
	"code_base/internal/domain/entity"
	"code_base/pkg/lib"
	"context"
)

type ProductRepository interface {
	lib.Closable
	Create(context.Context, entity.Product) (string, error)
	FindByID(context.Context, string) (entity.Product, error)
	FindByName(context.Context, string) (entity.Product, error)
	GetAll(context.Context) ([]entity.Product, error)
}
