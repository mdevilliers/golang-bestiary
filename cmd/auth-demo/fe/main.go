package main

import (
	"fmt"
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	auth "github.com/mdevilliers/golang-bestiary/pkg/auth-token"
	pb "github.com/mdevilliers/golang-bestiary/cmd/auth-demo/proto"
)

const (
	svc1address    = "localhost:50051"
	svc1servername = "waterzooi.test.google.be"
)

func main() {

	fmt.Println("fe svc running...")

	// create an authentication token
	signer, err := auth.NewTokenSigner("../keys/auth.rsa")

	if err != nil {
		log.Fatal(err)
		return
	}
	
	token := auth.NewToken(
		map[string]interface{}{
    		"role1": true,
    		"role2": true,
    		"role3": true,
    		"role4": true,
		},
	)
	
	// sign the token
	signedToken, err := signer.Sign(token)

	if err != nil {
		log.Fatal(err)
		return
	}

	// construct a context with the signed token
	ctxMarshaller := auth.NewContextMarshaller(context.Background())
	ctx := ctxMarshaller.Marshal(signedToken)
	
	if err != nil {
		log.Fatal(err)
		return
	}

	// make remote call over TLS
	creds, err := credentials.NewClientTLSFromFile("../keys/ca.pem", svc1servername)

	if err != nil {
		log.Fatalf("error loading creds: %v", err)
	}

	conn, err := grpc.Dial(svc1address, grpc.WithTransportCredentials(creds))

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "world"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.Message)
}
