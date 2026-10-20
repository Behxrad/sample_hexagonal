package application

import (
	inbound_api "code_base/internal/adapter/inbound/api"
	"code_base/internal/adapter/outbound/repository"
	"code_base/internal/service"
	"code_base/pkg/lib"
	"context"
)

type CommerceApp struct {
	httpServer lib.Server
	grpcServer lib.Server
	stop       func() error
}

func NewCommerceApp(ctx context.Context) CommerceApp {
	productRepository := repository.NewMongoProductRepository(ctx)
	productService := service.NewProductService(productRepository)
	httpServer := inbound_api.NewHttpServer(productService)
	grpcServer := inbound_api.NewGRPCServer(productService)

	return CommerceApp{
		httpServer: httpServer,
		grpcServer: grpcServer,
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
			return nil
		},
	}
}

func (c CommerceApp) Start(ctx context.Context) error {
	err := c.httpServer.Serve(ctx)
	if err != nil {
		return err
	}
	err = c.grpcServer.Serve(ctx)
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
