package repository

import (
	"context"

	"github.com/Behxrad/sample_hexagonal/internal/domain/entity"

	"192.168.100.6/coreinno/common.git/lib"
)

type ProductRepository interface {
	lib.Closable
	Create(context.Context, entity.Product) (string, error)
	FindByID(context.Context, string) (entity.Product, error)
	FindByName(context.Context, string) (entity.Product, error)
	GetAll(context.Context) ([]entity.Product, error)
}
