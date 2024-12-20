package main

import (
	"context"
	"io"
	"log"
	"time"

	pb "github.com/Vibhanshu17/Go/go_grpc/coffeeshop_proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:9001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server")
	}
	defer conn.Close()

	coffeeShopClient := pb.NewCoffeeShopClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	menuStream, err := coffeeShopClient.GetMenu(ctx, &pb.MenuRequest{})
	for err != nil {
		log.Fatalf("Error in calling GetMenu: %v", err)
		menuStream, err = coffeeShopClient.GetMenu(ctx, &pb.MenuRequest{})
	}
	done := make(chan bool)
	var items []*pb.Item

	go func() {
		for {
			resp, err := menuStream.Recv()
			if err == io.EOF {
				done <- true
				return
			}
			if err != nil {
				log.Fatalf("Error in Receiving GetMenu stream: %v", err)
			}
			items = resp.Items
			log.Printf("response received: %v", resp.Items)
		}
	}()
	<-done

	receipt, err := coffeeShopClient.PlaceOrder(ctx, &pb.Order{Items: items})
	if err != nil {
		log.Fatalf("Error in PlaceOrder: %v", err)
	}
	log.Printf("received receipt: %v", receipt)

	status, err := coffeeShopClient.GetOrderStatus(ctx, receipt)
	if err != nil {
		log.Fatalf("Error in GetOrderStatus: %v", err)
	}
	log.Printf("received order status: %v", status)
}
