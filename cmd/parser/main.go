package main

import (
	"log"
	"net"
	"os"

	pb "github.com/ginkgo1981/entries-parser/api/cotaparser/v1"
	"github.com/ginkgo1981/entries-parser/server"
	"google.golang.org/grpc"
)

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}

	s := grpc.NewServer()
	pb.RegisterCotaServer(s, server.Cota{})

	log.Println("start server on: ", port)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v\n", err)
	}
}
