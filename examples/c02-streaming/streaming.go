package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"maps"
	"slices"
	"time"

	"google.golang.org/grpc"

	pb "github.com/andrewmassart/go-learn-grpc/catalog"
	"github.com/andrewmassart/go-learn-grpc/internal/catalog"
)

type catalogServer struct {
	pb.UnimplementedCatalogServiceServer
	products   map[string]*pb.Product
	quantities map[string]int64
}

func (server *catalogServer) ListProducts(request *pb.ListProductsRequest, stream grpc.ServerStreamingServer[pb.ListProductsResponse]) error {
	for _, sku := range slices.Sorted(maps.Keys(server.products)) {
		if err := stream.Send(&pb.ListProductsResponse{Product: server.products[sku]}); err != nil {
			return err
		}
	}
	return nil
}

func (server *catalogServer) UpdateInventory(stream grpc.ClientStreamingServer[pb.UpdateInventoryRequest, pb.UpdateInventoryResponse]) error {
	var totalUnits int64
	for {
		request, err := stream.Recv()
		if err == io.EOF { // client finished sending
			return stream.SendAndClose(&pb.UpdateInventoryResponse{TotalUnits: totalUnits})
		}
		if err != nil {
			return err
		}
		server.quantities[request.Sku] += request.Quantity
		totalUnits += request.Quantity
	}
}

func (server *catalogServer) CheckStock(stream grpc.BidiStreamingServer[pb.CheckStockRequest, pb.CheckStockResponse]) error {
	for {
		request, err := stream.Recv()
		if err == io.EOF { // client finished asking
			return nil
		}
		if err != nil {
			return err
		}
		response := &pb.CheckStockResponse{Sku: request.Sku, Quantity: server.quantities[request.Sku]}
		if err := stream.Send(response); err != nil {
			return err
		}
	}
}

func listProducts(ctx context.Context, client pb.CatalogServiceClient) {
	stream, err := client.ListProducts(ctx, &pb.ListProductsRequest{})
	if err != nil {
		log.Fatalf("client.ListProducts failed: %v", err)
	}
	for {
		response, err := stream.Recv()
		if err == io.EOF { // server finished sending
			return
		}
		if err != nil {
			log.Fatalf("client.ListProducts: stream.Recv failed: %v", err)
		}
		fmt.Println("in catalog:", response.Product.Sku, response.Product.Name)
	}
}

func updateInventory(ctx context.Context, client pb.CatalogServiceClient) {
	stream, err := client.UpdateInventory(ctx)
	if err != nil {
		log.Fatalf("client.UpdateInventory failed: %v", err)
	}
	for _, request := range []*pb.UpdateInventoryRequest{
		{Sku: "SHOE-001", Quantity: 40},
		{Sku: "SHOE-002", Quantity: 25},
	} {
		fmt.Println("delivering:", request.Sku, request.Quantity)
		if err := stream.Send(request); err != nil {
			log.Fatalf("client.UpdateInventory: stream.Send failed: %v", err)
		}
	}
	response, err := stream.CloseAndRecv() // done sending, wait for the one reply
	if err != nil {
		log.Fatalf("client.UpdateInventory: stream.CloseAndRecv failed: %v", err)
	}
	fmt.Println("shipment received:", response.TotalUnits, "units")
}

func checkStock(ctx context.Context, client pb.CatalogServiceClient) {
	stream, err := client.CheckStock(ctx)
	if err != nil {
		log.Fatalf("client.CheckStock failed: %v", err)
	}
	for _, sku := range []string{"SHOE-001", "SHOE-002"} {
		fmt.Println("checking:", sku)
		if err := stream.Send(&pb.CheckStockRequest{Sku: sku}); err != nil {
			log.Fatalf("client.CheckStock: stream.Send failed: %v", err)
		}
		response, err := stream.Recv()
		if err != nil {
			log.Fatalf("client.CheckStock: stream.Recv failed: %v", err)
		}
		fmt.Println("in stock:", response.Sku, response.Quantity)
	}
	if err := stream.CloseSend(); err != nil {
		log.Fatalf("client.CheckStock: stream.CloseSend failed: %v", err)
	}
}

func main() {
	client := catalog.Start(&catalogServer{
		products: map[string]*pb.Product{
			"SHOE-001": {Sku: "SHOE-001", Name: "trail runner", PriceCents: 12999},
			"SHOE-002": {Sku: "SHOE-002", Name: "road racer", PriceCents: 15999},
		},
		quantities: map[string]int64{},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("-- server streaming: ListProducts --")
	listProducts(ctx, client)

	fmt.Println("\n-- client streaming: UpdateInventory --")
	updateInventory(ctx, client)

	fmt.Println("\n-- bidirectional streaming: CheckStock --")
	checkStock(ctx, client)
}
