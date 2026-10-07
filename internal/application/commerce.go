package application

import (
	inbound_api "code_base/internal/adapter/inbound/api"
	"code_base/internal/adapter/outbound/repository"
	"code_base/internal/service"
)

// Composition root
type CommerceApp struct {
	httpServer inbound_api.HttpServer
}

func NewCommerceApp() CommerceApp {
	productRepository := repository.NewPostgresProductRepository()
	productService := service.NewProductService(productRepository)
	server := inbound_api.NewHttpServer(productService)

	return CommerceApp{
		httpServer: server,
	}
}

func (c CommerceApp) Start() error {
	c.httpServer.Start()
	return nil
}
