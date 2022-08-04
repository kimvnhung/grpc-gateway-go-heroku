package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	_ "github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	pb "hung.kv/protogenerated"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	grpcServerEndpoint = flag.String("server-port", "1998", "gRPC server port")
	grpcGatewayPort    = flag.String("gateway-port", "1997", "Gateway port")
)

type server struct {
	pb.UnimplementedSimulatorServer
}

func (s *server) Home(ctx context.Context, in *pb.Empty) (*pb.HelloReply, error) {
	log.Printf("Home")
	return &pb.HelloReply{Message: "Hello World"}, nil
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	flag.Parse()
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	*grpcGatewayPort = port

	err := establishGrpcServer()

	if err != nil {
		log.Fatal(err)
	}

	err = establisGrpcGateway()

	if err != nil {
		log.Fatal(err)
	}

}

func establishGrpcServer() error {
	// Create a listener on TCP port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *grpcServerEndpoint))
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Create a gRPC server object
	s := grpc.NewServer()
	// Attach the Greeter service to the server
	pb.RegisterSimulatorServer(s, &server{})
	// Serve gRPC server
	log.Printf("server listening at %v", lis.Addr())
	go func() {
		log.Fatalln(s.Serve(lis))
	}()
	return nil
}

func establisGrpcGateway() error {
	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	conn, err := grpc.DialContext(
		context.Background(),
		fmt.Sprintf("0.0.0.0:%s", *grpcServerEndpoint),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()
	// Register Greeter
	err = pb.RegisterSimulatorHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	gwServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", *grpcGatewayPort),
		Handler: gwmux,
	}

	log.Printf("Serving gRPC-Gateway on http://0.0.0.0:%s\n", *grpcGatewayPort)
	log.Fatalln(gwServer.ListenAndServe())

	return nil
}
