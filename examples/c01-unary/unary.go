package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/andrewmassart/go-learn-grpc/catalog"
	"github.com/andrewmassart/go-learn-grpc/internal/catalog"
)

type catalogServer struct {
	pb.UnimplementedCatalogServiceServer
	products map[string]*pb.Product
}

func (server *catalogServer) GetProduct(ctx context.Context, request *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	product, found := server.products[request.Sku]
	if !found {
		return nil, fmt.Errorf("no product with sku %q", request.Sku)
	}
	return &pb.GetProductResponse{Product: product}, nil
}

func main() {
	client := catalog.Start(&catalogServer{products: map[string]*pb.Product{
		"SHOE-001": {Sku: "SHOE-001", Name: "trail runner", PriceCents: 12999},
		"SHOE-002": {Sku: "SHOE-002", Name: "road racer", PriceCents: 15999},
	}})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	sku := "SHOE-001"
	fmt.Println("requesting:", sku)
	response, err := client.GetProduct(ctx, &pb.GetProductRequest{Sku: sku})
	if err != nil {
		log.Fatalf("client.GetProduct failed: %v", err)
	}
	product := response.Product
	fmt.Println("received:", product.Sku, product.Name, product.PriceCents)
}
