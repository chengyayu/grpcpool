package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	pool "github.com/chengyayu/grpcpool"
	"github.com/chengyayu/grpcpool/example/pb"
)

var addr = flag.String("addr", "127.0.0.1:40000", "the address to connect to")

func main() {
	flag.Parse()

	p, err := pool.New(*addr)
	if err != nil {
		log.Fatalf("failed to new pool: %v", err)
	}
	defer p.Close()

	conn, err := p.Get()
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
