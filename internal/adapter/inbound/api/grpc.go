package api

import (
	"context"
	"fmt"
	"log"
	"net"

	"192.168.100.6/coreinno/proto_common.git/gen/sample/inventory/v1"
	inbound_headers "192.168.100.6/coreinno/proto_common.git/model/common/v1"
	"github.com/Behxrad/sample_hexagonal/configs"
	"github.com/Behxrad/sample_hexagonal/internal/port/inbound/api"

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
	r, err := s.ps.AddProduct(ctx, api.AddProductCommand{
		MetaData: inbound_headers.FromProto(cmd.Metadata),
		Product: api.Product{
			Count: uint16(cmd.Count), //nolint
			Price: uint32(cmd.Price), //nolint
			Name:  cmd.Name,
		}})
	if err != nil {
		return nil, err
	}
	return &inventory.CreateProductResult{
		Metadata: r.MetaData.ToProto(),
		Id:       r.ID,
	}, nil
}

func (s GRPCServer) GetProducts(ctx context.Context, qry *inventory.GetProductsQuery) (*inventory.GetProductsResult, error) {
	r, err := s.ps.ListProducts(ctx, api.ListProductsQuery{
		MetaData: inbound_headers.FromProto(qry.Metadata),
	})
	if err != nil {
		return nil, err
	}

	return &inventory.GetProductsResult{
		Metadata: r.MetaData.ToProto(),
		Products: func() []*inventory.Product {
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
