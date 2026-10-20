package repository

import (
	"code_base/internal/domain/entity"
	"code_base/internal/port/outbound/repository"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type mongoLogRepository struct {
	Client *mongo.Client
}

func NewMongoProductRepository(ctx context.Context) repository.ProductRepository {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().
		ApplyURI("mongodb://app_user:app_password@localhost:27017/inventory").
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(30 * time.Minute)

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	return mongoLogRepository{
		Client: client,
	}
}

func (m mongoLogRepository) Close(ctx context.Context) error {
	return m.Client.Disconnect(ctx)
}

func (m mongoLogRepository) Create(ctx context.Context, product entity.Product) (string, error) {
	result, err := m.Client.
		Database("inventory").
		Collection("products").
		InsertOne(ctx, product)
	if err != nil {
		return "", err
	}

	id, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		return "", errors.New("unexpected MongoDB inserted ID type")
	}
	return id.Hex(), nil
}

type productDocument struct {
	ID      bson.ObjectID  `bson:"_id"`
	Product entity.Product `bson:",inline"`
}

func (m mongoLogRepository) FindByID(ctx context.Context, id string) (entity.Product, error) {
	var p productDocument

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return entity.Product{}, err
	}

	err = m.Client.
		Database("inventory").
		Collection("products").
		FindOne(ctx, bson.D{
			{Key: "_id", Value: objectID},
		}).
		Decode(&p)
	if err != nil {
		return entity.Product{}, err
	}

	p.Product.ID = p.ID.Hex()
	return p.Product, nil
}

func (m mongoLogRepository) FindByName(ctx context.Context, name string) (entity.Product, error) {
	var p productDocument

	err := m.Client.
		Database("inventory").
		Collection("products").
		FindOne(ctx, bson.D{
			{Key: "name", Value: name},
		}).Decode(&p)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return entity.Product{}, nil
		}
		return entity.Product{}, err
	}

	p.Product.ID = p.ID.Hex()
	return p.Product, nil
}

func (m mongoLogRepository) GetAll(ctx context.Context) ([]entity.Product, error) {
	products := make([]entity.Product, 0)

	cursor, err := m.Client.
		Database("inventory").
		Collection("products").
		Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var p productDocument
		if err := cursor.Decode(&p); err != nil {
			return nil, err
		}

		p.Product.ID = p.ID.Hex()
		products = append(products, p.Product)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return products, nil
}
