package catalog

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/andrewmassart/go-learn-grpc/catalog"
)

type Server struct {
	pb.UnimplementedCatalogServiceServer
	products map[string]*pb.Product
}

func NewServer() *Server {
	return &Server{products: map[string]*pb.Product{
		"SHOE-001": {Sku: "SHOE-001", Name: "trail runner", PriceCents: 12999},
		"SHOE-002": {Sku: "SHOE-002", Name: "road racer", PriceCents: 15999},
	}}
}

func (server *Server) GetProduct(ctx context.Context, request *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	product, found := server.products[request.Sku]
	if !found {
		return nil, fmt.Errorf("no product with sku %q", request.Sku)
	}
	return &pb.GetProductResponse{Product: product}, nil
}

func Start() pb.CatalogServiceClient {
	listener, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCatalogServiceServer(grpcServer, NewServer())
	go grpcServer.Serve(listener) // serve in the background, the caller acts as the client

	connection, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	return pb.NewCatalogServiceClient(connection)
}
