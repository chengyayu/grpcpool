package main

import (
	"context"
	"flag"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	pool "github.com/chengyayu/grpcpool"
	"github.com/chengyayu/grpcpool/example/pb"
)

var addr = flag.String("addr", "0.0.0.0:30000", "port number")

// server implements EchoServer.
type server struct {
	pb.UnimplementedEchoServer
}

func (s *server) Say(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	log.Printf("server replay: %+v", req)
	return &pb.EchoResponse{Message: req.GetMessage()}, nil
}

func main() {
	flag.Parse()

	listen, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.InitialWindowSize(pool.InitialWindowSize),
		grpc.InitialConnWindowSize(pool.InitialConnWindowSize),
		grpc.MaxSendMsgSize(pool.MaxSendMsgSize),
		grpc.MaxRecvMsgSize(pool.MaxRecvMsgSize),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			PermitWithoutStream: true,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    pool.KeepAliveTime,
			Timeout: pool.KeepAliveTimeout,
		}),
	)
	pb.RegisterEchoServer(s, &server{})

	if err := s.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
