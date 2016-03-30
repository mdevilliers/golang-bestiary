package main

import (
	"errors"
	"fmt"
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/mdevilliers/golang-bestiary/cmd/auth-demo/proto"
	auth "github.com/mdevilliers/golang-bestiary/pkg/auth-token"
	"google.golang.org/grpc/credentials"
)

const (
	authServerAddress = "localhost:50000"
	port              = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {

	// this should be done as an interceptor
	// and cached on a rotate schedule
	authServiceRequestURL := fmt.Sprintf("http://%s/publickey", authServerAddress)
	authority, err := auth.NewContextAuthorityFromURL(authServiceRequestURL, ctx)

	if err != nil {
		log.Println(err)
		return nil, errors.New("Go away!")
	}

	ok, err := authority.IsInRole("role1")

	if !ok {
		log.Println(ok, err)
		return nil, errors.New("Go away!")
	}

	fmt.Println("HelloRequest in svc1")
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func main() {

	fmt.Println("svc1 running....")

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	creds, err := credentials.NewServerTLSFromFile("../keys/server1.pem", "../keys/server1.key")

	if err != nil {
		log.Fatal("failed to load credentials", err)
	}
	opts := []grpc.ServerOption{grpc.Creds(creds)}
	s := grpc.NewServer(opts...)
	pb.RegisterGreeterServer(s, &server{})
	s.Serve(lis)
}
