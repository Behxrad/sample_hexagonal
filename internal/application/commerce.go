package application

import (
	inbound_api "code_base/internal/adapter/inbound/api"
	"code_base/internal/adapter/outbound/repository"
	"code_base/internal/service"
	"context"
)

// Composition root
type CommerceApp struct {
	httpServer inbound_api.HttpServer
	stop       func() error
}

func NewCommerceApp(ctx context.Context) CommerceApp {
	productRepository := repository.NewPostgresProductRepository(ctx)
	productService := service.NewProductService(productRepository)
	server := inbound_api.NewHttpServer(productService)

	return CommerceApp{
		httpServer: server,
		stop: func() error {
			err := productRepository.Close(ctx)
			if err != nil {
				return err
			}
			err = server.Close(ctx)
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func (c CommerceApp) Start(ctx context.Context) error {
	err := c.httpServer.Serve()
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
