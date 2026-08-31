package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/andrewmassart/go-learn-grpc/catalog"
	"github.com/andrewmassart/go-learn-grpc/internal/catalog"
)

type catalogServer struct {
	pb.UnimplementedCatalogServiceServer
	products map[string]*pb.Product
}

func (server *catalogServer) GetProduct(ctx context.Context, request *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	if request.Sku == "" {
		return nil, status.Error(codes.InvalidArgument, "sku is required")
	}

	select {
	case <-time.After(150 * time.Millisecond):
	case <-ctx.Done():
		return nil, status.FromContextError(ctx.Err()).Err()
	}

	product, found := server.products[request.Sku]
	if !found {
		return nil, status.Errorf(codes.NotFound, "no product with sku %q", request.Sku)
	}
	return &pb.GetProductResponse{Product: product}, nil
}

func getProduct(client pb.CatalogServiceClient, sku string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fmt.Printf("requesting %q with %v timeout\n", sku, timeout)
	response, err := client.GetProduct(ctx, &pb.GetProductRequest{Sku: sku})
	if err != nil {
		grpcStatus, ok := status.FromError(err)
		if !ok {
			log.Fatalf("client.GetProduct error: %v", err)
		}
		fmt.Println("error:", grpcStatus.Code(), "-", grpcStatus.Message())
		return
	}
	fmt.Println("received:", response.Product.Sku, response.Product.Name, response.Product.PriceCents)
}

func main() {
	client := catalog.Start(&catalogServer{products: map[string]*pb.Product{
		"SHOE-001": {Sku: "SHOE-001", Name: "trail runner", PriceCents: 12999},
		"SHOE-002": {Sku: "SHOE-002", Name: "road racer", PriceCents: 15999},
	}})

	fmt.Println("-- ok --")
	getProduct(client, "SHOE-001", time.Second)

	fmt.Println("\n-- invalid argument --")
	getProduct(client, "", time.Second)

	fmt.Println("\n-- not found --")
	getProduct(client, "SHOE-999", time.Second)

	fmt.Println("\n-- deadline exceeded --")
	getProduct(client, "SHOE-001", 50*time.Millisecond)
}
