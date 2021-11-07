package efremBackend

import (
	"context"
	"fmt"
	"github.com/iamrajiv/helloworld-grpc-gateway/proto"
	"github.com/iamrajiv/helloworld-grpc-gateway/proto/auth"
	"google.golang.org/grpc/metadata"
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pbAuth "github.com/iamrajiv/helloworld-grpc-gateway/proto/auth"
	pbHelloWorld "github.com/iamrajiv/helloworld-grpc-gateway/proto/helloworld"
)

type server struct{}

func NewServer() *server {
	return &server{}
}

type AuthServer struct {
}

func (s *AuthServer) AuthHandler(c context.Context, in *pbAuth.AuthRequest) (*pbAuth.HelloNewReply, error) {
	md, ok := metadata.FromIncomingContext(c)
	r := c.Value("test2")
	fmt.Println(md, ok, r)

	return &pbAuth.HelloNewReply{Message: in.Auth + " auth"}, nil
}

func (*server) SayHello(c context.Context, in *pbHelloWorld.HelloRequest) (*pbHelloWorld.HelloReply, error) {
	md, ok := metadata.FromIncomingContext(c)
	fmt.Println(md, ok)
	return &pbHelloWorld.HelloReply{Message: in.Name + in.Test + " world"}, nil
}

func newGateway(ctx context.Context, conn *grpc.ClientConn, opts []gwruntime.ServeMuxOption) (http.Handler, error) {

	mux := gwruntime.NewServeMux(opts...)
	err := pbHelloWorld.RegisterGreeterHandler(ctx, mux, conn)
	err = pbAuth.RegisterAuthHandler(ctx, mux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}
	return mux, nil
}

type Options struct {
	// Addr is the address to listen
	Addr string

	// OpenAPIDir is a path to a directory from which the server
	// serves OpenAPI specs.
	OpenAPIDir string

	// Mux is a list of options to be passed to the gRPC-Gateway multiplexer
	Mux []gwruntime.ServeMuxOption
}

func main() {
	opts := Options{}
	opts.Mux = []gwruntime.ServeMuxOption{gwruntime.WithMetadata(proto.Test)}
	// Create a listener on TCP port
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Create a gRPC server object
	s := grpc.NewServer()
	srv := &server{}

	// Attach the Greeter service to the server
	pbHelloWorld.RegisterGreeterServer(s, srv)
	auth.RegisterAuthServer(s, &AuthServer{})

	// Serve gRPC server
	log.Println("Serving gRPC on 0.0.0.0:8080")
	go func() {
		log.Fatalln(s.Serve(lis))
	}()

	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	conn, err := grpc.DialContext(
		context.Background(),
		"0.0.0.0:8080",
		grpc.WithBlock(),
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	mux := http.NewServeMux()

	gw, err := newGateway(context.Background(), conn, opts.Mux)
	if err != nil {
		fmt.Println(err)
	}
	mux.Handle("/", gw)

	gwServer := &http.Server{
		Addr:    ":8090",
		Handler: proto.AllowCORS(mux),
	}

	log.Println("Serving gRPC-Gateway on http://0.0.0.0:8090")
	log.Fatalln(gwServer.ListenAndServe())
}
