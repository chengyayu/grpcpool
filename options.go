package grpcpool

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"time"
)

const (
	// DialTimeout the timeout of create connection
	DialTimeout = 5 * time.Second

	// BackoffMaxDelay provided maximum delay when backing off after failed connection attempts.
	BackoffMaxDelay = 3 * time.Second

	// KeepAliveTime 如此时长后客户端未收到响应则会 ping 服务端。
	KeepAliveTime = time.Duration(10) * time.Second

	// KeepAliveTimeout 客户端 ping 服务端后，如此时长没响应会关闭连接。
	KeepAliveTimeout = time.Duration(3) * time.Second

	// InitialWindowSize we set it 1GB is to provide system's throughput.
	InitialWindowSize = 1 << 30

	// InitialConnWindowSize we set it 1GB is to provide system's throughput.
	InitialConnWindowSize = 1 << 30

	// MaxSendMsgSize set max gRPC request message size sent to server.
	// If any request message size is larger than current value, an error will be reported from gRPC.
	MaxSendMsgSize = 4 << 30

	// MaxRecvMsgSize set max gRPC receive message size received from server.
	// If any message size is larger than current value, an error will be reported from gRPC.
	MaxRecvMsgSize = 4 << 30
)

const (
	// DftMaxIdle see options.MaxIdle
	DftMaxIdle = int(8)
	// DftMaxActive see options.MaxActive
	DftMaxActive = int(64)
	// DftMaxConcurrentStreams see options.MaxConcurrentStreams
	DftMaxConcurrentStreams = int(64)
)

// Option is an options setting function.
type Option func(o *options)

// options are params for creating grpc connect pool.
type options struct {
	// dial is an application supplied function for creating and configuring a connection.
	dial func(address string) (*grpc.ClientConn, error)

	// maxIdle is a maximum number of idle connections in the pool.
	maxIdle int

	// maxActive is a maximum number of connections allocated by the pool at a given time.
	// When zero, there is no limit on the number of connections in the pool.
	maxActive int

	// maxConcurrentStreams limit on the number of concurrent streams to each single connection
	maxConcurrentStreams int

	// If reuse is true and the pool is at the MaxActive limit, then Get() reuse
	// the connection to return, If reuse is false and the pool is at the maxActive limit,
	// create a one-time connection to return.
	reuse bool
}

// Dial with factory function for *grpc.ClientConn
func Dial(factoryFn func(address string) (*grpc.ClientConn, error)) Option {
	return func(o *options) { o.dial = factoryFn }
}

// MaxIdle with pool maxIdle
func MaxIdle(maxIdle int) Option {
	return func(o *options) { o.maxIdle = maxIdle }
}

// MaxActive with pool maxActive
func MaxActive(maxActive int) Option {
	return func(o *options) { o.maxActive = maxActive }
}

// MaxConcurrentStreams with pool maxConcurrentStreams
func MaxConcurrentStreams(maxConcurrentStreams int) Option {
	return func(o *options) { o.maxConcurrentStreams = maxConcurrentStreams }
}

// Reuse with pool reuse
func Reuse(reuse bool) Option {
	return func(o *options) { o.reuse = reuse }
}

// DftDial return a grpc connection with defined configurations.
func DftDial(address string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DialTimeout)
	defer cancel()
	return grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBackoffMaxDelay(BackoffMaxDelay),
		grpc.WithInitialWindowSize(InitialWindowSize),
		grpc.WithInitialConnWindowSize(InitialConnWindowSize),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(MaxSendMsgSize), grpc.MaxCallRecvMsgSize(MaxRecvMsgSize)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                KeepAliveTime,
			Timeout:             KeepAliveTimeout,
			PermitWithoutStream: true,
		}))
}
