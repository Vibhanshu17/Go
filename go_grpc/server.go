package main

import (
	"context"
	"log"
	"net"

	pb "github.com/Vibhanshu17/Go/go_grpc/coffeeshop_proto"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedCoffeeShopServer
}

func (s *Server) GetMenu(menuRequest *pb.MenuRequest, server pb.CoffeeShop_GetMenuServer) error {
	log.Printf("GetMenu() called")
	items := []*pb.Item{
		{
			Id:   "1",
			Name: "Black Coffee",
		},
		{
			Id:   "2",
			Name: "Americano",
		},
		{
			Id:   "3",
			Name: "Vanilla Soy Chai Latte",
		},
	}

	// for i := range items {
	// 	server.Send(&pb.Menu{
	// 		Items: items[0 : i+1],
	// 	})
	// }
	log.Printf("Streaming menu items starting...")
	for _, item := range items {
		if err := server.Send(&pb.Menu{Items: []*pb.Item{item}}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) PlaceOrder(context.Context, *pb.Order) (*pb.Receipt, error) {
	return &pb.Receipt{
		Id: "ABC123",
	}, nil
}

func (s *Server) GetOrderStatus(context context.Context, receipt *pb.Receipt) (*pb.OrderStatus, error) {
	return &pb.OrderStatus{
		OrderId: receipt.Id,
		Status:  "In Progress",
	}, nil
}

func main() {
	log.Printf("main running")
	lis, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("listening on port: 9001")
	grpcServer := grpc.NewServer()

	log.Printf("new coffeeShopServer created")
	pb.RegisterCoffeeShopServer(grpcServer, &Server{}) // bind grpcServer to our Server struct

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %s", err)
	}
	log.Printf("new coffeeShopServer closed")
}
