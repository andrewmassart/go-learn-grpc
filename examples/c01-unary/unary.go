package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/andrewmassart/go-learn-grpc/catalog"
	"github.com/andrewmassart/go-learn-grpc/internal/catalog"
)

func main() {
	client := catalog.Start()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	response, err := client.GetProduct(ctx, &pb.GetProductRequest{Sku: "SHOE-001"})
	fmt.Println("requesting sku:", "SHOE-001")
	if err != nil {
		log.Fatal(err)
	}

	product := response.Product
	fmt.Println("received:", product.Sku, product.Name, product.PriceCents)
}
