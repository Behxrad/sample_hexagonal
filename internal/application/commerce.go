package application

import (
	"context"

	inbound_api "github.com/Behxrad/sample_hexagonal/internal/adapter/inbound/api"
	"github.com/Behxrad/sample_hexagonal/internal/adapter/inbound/event"
	"github.com/Behxrad/sample_hexagonal/internal/adapter/outbound/repository"
	"github.com/Behxrad/sample_hexagonal/internal/service"
)

type CommerceApp struct {
	start func() error
	stop  func() error
}

func NewCommerceApp(ctx context.Context) CommerceApp {
	productRepository := repository.NewMongoProductRepository(ctx)
	productService := service.NewProductService(productRepository)
	httpServer := inbound_api.NewHttpServer(productService)
	grpcServer := inbound_api.NewGRPCServer(productService)
	kafka := event.NewKafkaAdapter()

	return CommerceApp{
		start: func() error {
			err := httpServer.Serve(ctx)
			if err != nil {
				return err
			}
			err = grpcServer.Serve(ctx)
			if err != nil {
				return err
			}
			err = kafka.Start(ctx)
			if err != nil {
				return err
			}
			return nil
		},
		stop: func() error {
			err := productRepository.Close(ctx)
			if err != nil {
				return err
			}
			err = httpServer.Close(ctx)
			if err != nil {
				return err
			}
			err = grpcServer.Close(ctx)
			if err != nil {
				return err
			}
			err = kafka.Close()
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func (c CommerceApp) Start(ctx context.Context) error {
	err := c.start()
	if err != nil {
		return err
	}
	return nil
}

func (c CommerceApp) Stop(ctx context.Context) error {
	err := c.stop()
	if err != nil {
		return err
	}
	return nil
}
