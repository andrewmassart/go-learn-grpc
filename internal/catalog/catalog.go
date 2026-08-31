package catalog

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/andrewmassart/go-learn-grpc/catalog"
)

func Start(server pb.CatalogServiceServer, options ...grpc.ServerOption) pb.CatalogServiceClient {
	listener, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("net.Listen failed: %v", err)
	}

	grpcServer := grpc.NewServer(options...)
	pb.RegisterCatalogServiceServer(grpcServer, server)
	go func() { // serve in the background, the caller acts as the client
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("grpcServer.Serve failed: %v", err)
		}
	}()

	connection, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("grpc.NewClient failed: %v", err)
	}
	return pb.NewCatalogServiceClient(connection)
}
