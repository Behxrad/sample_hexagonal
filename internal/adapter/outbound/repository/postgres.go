package repository

import (
	"code_base/internal/domain/entity"
	"code_base/internal/port/outbound/repository"
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresProductRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresProductRepository(ctx context.Context) repository.ProductRepository {
	config, err := pgxpool.ParseConfig("postgres://app_user:secret@localhost:5434/app_db")
	if err != nil {
		log.Fatal(err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatal(err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		log.Fatal(err)
	}

	return postgresProductRepository{
		pool: pool,
	}
}

func (p postgresProductRepository) Close(ctx context.Context) error {
	p.pool.Close()
	log.Println("postgres closed")
	return nil
}

func (p postgresProductRepository) Create(ctx context.Context, product entity.Product) (string, error) {
	err := p.pool.QueryRow(ctx, `
        INSERT INTO inventory.products (name, count, price)
        VALUES ($1, $2, $3)
        RETURNING id
    `,
		product.Name,
		product.Count,
		product.Price,
	).Scan(&product.ID)

	if err != nil {
		return "", err
	}

	return product.ID, nil
}

func (p postgresProductRepository) FindByID(ctx context.Context, id string) (entity.Product, error) {
	var product entity.Product

	pid, err := strconv.Atoi(id)
	if err != nil {
		return entity.Product{}, err
	}

	err = p.pool.QueryRow(ctx, `
        SELECT id, name, count, price
        FROM inventory.products
        WHERE id = $1
    `, pid).Scan(
		&product.ID,
		&product.Name,
		&product.Count,
		&product.Price,
	)

	if err != nil {
		return entity.Product{}, err
	}

	return product, nil
}

func (p postgresProductRepository) FindByName(ctx context.Context, name string) (entity.Product, error) {
	var product entity.Product

	err := p.pool.QueryRow(ctx, `
        SELECT id, name, count, price
        FROM inventory.products
        WHERE name = $1
    `, name).Scan(
		&product.ID,
		&product.Name,
		&product.Count,
		&product.Price,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Product{}, nil
		}
		return entity.Product{}, err
	}

	return product, nil
}

func (p postgresProductRepository) GetAll(ctx context.Context) ([]entity.Product, error) {
	products := make([]entity.Product, 0)

	rows, err := p.pool.Query(ctx, `
        SELECT id, name, count, price
        FROM inventory.products
    `)
	if err != nil {
		return products, err
	}

	for rows.Next() {
		var product entity.Product
		err = rows.Scan(&product.ID, &product.Name, &product.Count, &product.Price)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return products, nil
			}
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}
