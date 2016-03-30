package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/mdevilliers/golang-bestiary/cmd/auth-demo/proto"
	auth "github.com/mdevilliers/golang-bestiary/pkg/auth-token"
)

const (
	authServerAddress = "localhost:50000"
	svc1address       = "localhost:50051"
	svc1servername    = "waterzooi.test.google.be"
)

func main() {

	fmt.Println("fe svc running...")

	log.Println("authenticating...")
	authServiceRequestURL := fmt.Sprintf("http://%s/authenticate", authServerAddress)

	resp, err := http.PostForm(authServiceRequestURL,
		url.Values{"username": {"mark"}, "password": {"secret"}})

	if err != nil {
		log.Fatal(err)
		return
	}

	defer resp.Body.Close()
	token, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println("calling service 1...")

	ctxMarshaller := auth.NewContextMarshaller(context.Background())
	ctx := ctxMarshaller.MarshalTrustedString(string(token))

	if err != nil {
		log.Fatal(err)
		return
	}

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
