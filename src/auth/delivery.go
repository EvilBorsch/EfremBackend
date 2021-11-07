package auth

import (
	"context"
	"fmt"
	"google.golang.org/grpc/metadata"
)

type AuthServer struct {
}

func (s *AuthServer) AuthHandler(c context.Context, in *pbAuth.AuthRequest) (*pbAuth.HelloNewReply, error) {
	md, ok := metadata.FromIncomingContext(c)
	r := c.Value("test2")
	fmt.Println(md, ok, r)

	return &pbAuth.HelloNewReply{Message: in.Auth + " auth"}, nil
}
