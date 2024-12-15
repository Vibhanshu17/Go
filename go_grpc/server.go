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

func (s *Server) GetMenu(menuRequest *pb.MenuRequest, server *pb.CoffeeShop_GetMenuServer) error {
	items := []*pb.Item{
		*pb.Item{
			Id:   "1",
			Name: "Black Coffee",
		},
		*pb.Item{
			Id:   "2",
			Name: "Americano",
		},
		*pb.Item{
			Id:   "3",
			Name: "Vanilla Soy Chai Latte",
		},
	}

	for i, _ := range items {
		server.Send(&pb.Menu{
			Items: items[0 : i+1],
		})
	}
	return nil
}

func (s *Server) PlaceOrder(context context.Context, Order *pb.Order) (*pb.Receipt, error) {
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
	lis, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterCoffeeShopServer(grpcServer, &Server{}) // bind grpcServer to our Server struct
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %s", err)
	}
}
