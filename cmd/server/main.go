package main

import (
	"log"
	"net"

	"github.com/AlexKomzzz/test_tages/pkg/api"
	libraryserver "github.com/AlexKomzzz/test_tages/pkg/server"
	"google.golang.org/grpc"
)

const (
	srvPort = ":8080"
)

func main() {

	lis, err := net.Listen("tcp", srvPort)
	if err != nil {
		log.Fatal("failed to listen: ", err)
	}

	srv := libraryserver.NewGRPCServer()
	s := grpc.NewServer()
	api.RegisterFileStorageServer(s, srv)

	if err := s.Serve(lis); err != nil {
		log.Fatal("failed to serve: ", err)
	}
}
