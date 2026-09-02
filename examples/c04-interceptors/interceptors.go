package main

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

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
		return nil, status.Errorf(codes.NotFound, "no product with sku %q", request.Sku)
	}
	return &pb.GetProductResponse{Product: product}, nil
}

func logCalls(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	response, err := handler(ctx, request) // run the rest of the chain, then the handler
	fmt.Println("logged:", info.FullMethod, status.Code(err), time.Since(start))
	return response, err
}

func requireToken(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	requestMetadata, _ := metadata.FromIncomingContext(ctx)
	tokens := requestMetadata.Get("authorization")
	if len(tokens) == 0 || tokens[0] != "secret-token" {
		return nil, status.Error(codes.Unauthenticated, "missing or invalid token")
	}
	return handler(ctx, request)
}

func getProduct(ctx context.Context, client pb.CatalogServiceClient, sku string) {
	response, err := client.GetProduct(ctx, &pb.GetProductRequest{Sku: sku})
	if err != nil {
		fmt.Println("error:", status.Code(err), "-", status.Convert(err).Message())
		return
	}
	fmt.Println("received:", response.Product.Sku, response.Product.Name)
}

func main() {
	client := catalog.Start(&catalogServer{products: map[string]*pb.Product{
		"SHOE-001": {Sku: "SHOE-001", Name: "trail runner", PriceCents: 12999},
		"SHOE-002": {Sku: "SHOE-002", Name: "road racer", PriceCents: 15999},
	}}, grpc.ChainUnaryInterceptor(logCalls, requireToken))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	fmt.Println("-- no token --")
	getProduct(ctx, client, "SHOE-001")

	fmt.Println("\n-- with token --")
	authCtx := metadata.AppendToOutgoingContext(ctx, "authorization", "secret-token")
	getProduct(authCtx, client, "SHOE-001")

	fmt.Println("\n-- with token, unknown sku --")
	getProduct(authCtx, client, "SHOE-999")
}
