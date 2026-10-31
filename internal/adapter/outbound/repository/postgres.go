package repository

import (
	"code_base/configs"
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
	connTimeout, err := time.ParseDuration(configs.ApplicationConfigs.DatabaseConnectTimeout)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(ctx, connTimeout)
	defer cancel()

	config, err := pgxpool.ParseConfig(configs.ApplicationConfigs.DatabaseURI)
	if err != nil {
		log.Fatal(err)
	}

	connIdleTime, err := time.ParseDuration(configs.ApplicationConfigs.DatabaseMaxConnIdleTime)
	if err != nil {
		log.Fatal(err)
	}

	connLifeTime, err := time.ParseDuration(configs.ApplicationConfigs.DatabaseMaxConnIdleTime)
	if err != nil {
		log.Fatal(err)
	}

	config.MaxConns = int32(configs.ApplicationConfigs.DatabaseMaxConnSize)
	config.MinConns = int32(configs.ApplicationConfigs.DatabaseMinConnSize)
	config.MaxConnLifetime = connLifeTime
	config.MaxConnIdleTime = connIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatal(err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		log.Fatal(errors.Join(errors.New("cannot connect to postgres"), err))
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
