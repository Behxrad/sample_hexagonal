package api

import (
	inventory "code_base/api/proto"
	"code_base/configs"
	"code_base/internal/port/inbound/api"
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	inventory.UnimplementedProductServiceServer
	gs *grpc.Server
	ps api.ProductService
}

func NewGRPCServer(service api.ProductService) GRPCServer {
	g := GRPCServer{
		ps: service,
	}
	sr := grpc.NewServer()
	inventory.RegisterProductServiceServer(sr, g)
	reflection.Register(sr)
	g.gs = sr
	return g
}

func (s GRPCServer) Serve(ctx context.Context) error {
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", configs.ApplicationConfigs.GRPCHost,
			configs.ApplicationConfigs.GRPCPort))
		if err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		if err := s.gs.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	return nil
}

func (s GRPCServer) Close(ctx context.Context) error {
	log.Println("grpc server closed")
	s.gs.GracefulStop()
	return nil
}

func (s GRPCServer) CreateProduct(ctx context.Context, cmd *inventory.CreateProductCommand) (*inventory.CreateProductResult, error) {
	r, err := s.ps.AddProduct(ctx, api.AddProductCommand{Product: api.Product{
		Count: uint16(cmd.Count),
		Price: uint32(cmd.Price),
		Name:  cmd.Name,
	}})
	if err != nil {
		return nil, err
	}
	return &inventory.CreateProductResult{Id: r.ID}, nil
}
func (s GRPCServer) GetProducts(ctx context.Context, qry *inventory.GetProductsQuery) (*inventory.GetProductsResult, error) {
	r, err := s.ps.ListProducts(ctx, api.ListProductsQuery{})
	if err != nil {
		return nil, err
	}

	return &inventory.GetProductsResult{Products: func() []*inventory.Product {
		products := make([]*inventory.Product, 0)
		for _, p := range r.Products {
			products = append(products, &inventory.Product{
				Id:    p.ID,
				Count: int32(p.Count),
				Price: int64(p.Price),
				Name:  p.Name,
			})
		}
		return products
	}()}, nil
}
