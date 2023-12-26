package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/sercand/kuberesolver/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"log"
	"time"

	pool "github.com/chengyayu/grpcpool"
	"github.com/chengyayu/grpcpool/example/pb"
)

var addr = flag.String("addr", "127.0.0.1:30000", "the address to connect to")

func main() {
	flag.Parse()

	var (
		namespace   = "default"
		servicename = "grpc-server"
		serviceport = 30000
	)
	kuberesolver.RegisterInCluster()
	endpoint := fmt.Sprintf("kubernetes:///%s.%s:%d", servicename, namespace, serviceport) // for kuberesolver

	dialFn := func(address string) (*grpc.ClientConn, error) {
		ctx, cancel := context.WithTimeout(context.Background(), pool.DialTimeout)
		defer cancel()
		return grpc.DialContext(ctx, address,
			grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`), // for kuberesolver
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBackoffMaxDelay(pool.BackoffMaxDelay),
			grpc.WithInitialWindowSize(pool.InitialWindowSize),
			grpc.WithInitialConnWindowSize(pool.InitialConnWindowSize),
			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(pool.MaxSendMsgSize)),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(pool.MaxRecvMsgSize)),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                pool.KeepAliveTime,
				Timeout:             pool.KeepAliveTimeout,
				PermitWithoutStream: true,
			}))
	}

	p, err := pool.New(endpoint, pool.Dial(dialFn), pool.MaxConcurrentStreams(4))
	if err != nil {
		log.Fatalf("failed to new pool: %v", err)
	}
	defer p.Close()

	loop(p)

	hang := make(chan int)
	<-hang
}

func loop(p pool.Pool) {
	defer holdpanic()
	do := func() {
		conn, err := p.Get()
		log.Printf("conn: %p, pool: %v", conn.Value(), p.Status())
		if err != nil {
			log.Fatalf("failed to get conn: %v", err)
		}
		defer conn.Close()

		client := pb.NewEchoClient(conn.Value())
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		res, err := client.Say(ctx, &pb.EchoRequest{Message: []byte("hi")})
		if err != nil {
			log.Fatalf("unexpected error from Say: %v", err)
		}
		fmt.Println("rpc response:", res)
	}
	for {
		do()
		time.Sleep(time.Second * 5)
	}
}

func holdpanic() {
	if err := recover(); err != nil {
		log.Printf("hold panic err: %+v", err)
	}
}
