// Copyright 2019 shimingyah. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// ee the License for the specific language governing permissions and
// limitations under the License.
//
// Modifications copyright (C) 2023 chengyayu.

package grpcpool

import (
	"context"
	"flag"
	"github.com/chengyayu/grpcpool/example/single/pb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var endpoint = flag.String("endpoint", "127.0.0.1:40000", "grpc server endpoint")

var (
	oncePoolOnce sync.Once
	oncePool     *pool
)

func newOncePool(opts ...Option) (*pool, error) {
	var (
		p   Pool
		err error
	)
	oncePoolOnce.Do(func() {
		p, err = New(*endpoint, opts...)
		log.Printf("***** new pool[%p],%s", p, p.Status())
		oncePool = p.(*pool)
	})

	return oncePool, err
}

func TestBadConn(t *testing.T) {
	opts := []Option{
		Dial(func(address string) (*grpc.ClientConn, error) {
			ctx, cancel := context.WithTimeout(context.Background(), DialTimeout)
			defer cancel()
			return grpc.DialContext(ctx, address,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				//grpc.WithConnectParams(grpc.ConnectParams{Backoff: backoff.Config{MaxDelay: BackoffMaxDelay}}),
				grpc.WithInitialWindowSize(InitialWindowSize),
				grpc.WithInitialConnWindowSize(InitialConnWindowSize),
				grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(MaxSendMsgSize)),
				grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MaxRecvMsgSize)),
				grpc.WithKeepaliveParams(keepalive.ClientParameters{
					Time:                KeepAliveTime,
					Timeout:             KeepAliveTimeout,
					PermitWithoutStream: true,
				}))
		}),
		MaxIdle(1),
		MaxActive(1),
		MaxConcurrentStreams(DftMaxConcurrentStreams),
		Reuse(true), // 池满不复用连接，创建新连接
	}

	f := func() {
		p, err := newOncePool(opts...)
		if err != nil {
			panic(err)
		}

		conn, err := p.Get()
		if err != nil {
			t.Fatalf("failed to get conn: %v", err)
		}
		defer conn.Close()

		t.Log(p.Status())
		client := pb.NewEchoClient(conn.Value())
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		data := make([]byte, size)
		t.Logf("1 %p, %s", conn.Value(), conn.Value().GetState().String())
		_, err = client.Say(ctx, &pb.EchoRequest{Message: data})
		t.Logf("2 %p, %s", conn.Value(), conn.Value().GetState().String())
		if err != nil {
			t.Logf("unexpected error from Say: %v", err)
		}
	}

	f()
	f()

	time.Sleep(30 * time.Second)

	f()
	f()
}

// DialTest return a simple grpc connection with defined configurations.
func DialTest(address string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DialTimeout)
	defer cancel()
	return grpc.DialContext(ctx, address, grpc.WithInsecure())
}

func newPool(opts ...Option) (Pool, *pool, error) {
	opts = append(opts, Dial(DialTest))
	p, err := New(*endpoint, opts...)
	return p, p.(*pool), err
}

func TestNew(t *testing.T) {
	p, nativePool, err := newPool()
	require.NoError(t, err)
	defer p.Close()

	options := nativePool.opt
	require.EqualValues(t, 0, nativePool.index)
	require.EqualValues(t, 0, nativePool.ref)
	require.EqualValues(t, options.maxIdle, nativePool.current)
	require.EqualValues(t, options.maxActive, len(nativePool.conns))
}

func TestNew2(t *testing.T) {
	_, err := New("")
	require.Error(t, err)

	_, err = New("127.0.0.1:8080", Dial(nil))
	require.Error(t, err)

	_, err = New("127.0.0.1:8080", MaxConcurrentStreams(0))
	require.Error(t, err)

	_, err = New("127.0.0.1:8080", MaxIdle(0))
	require.Error(t, err)

	_, err = New("127.0.0.1:8080", MaxActive(0))
	require.Error(t, err)

	_, err = New("127.0.0.1:8080", MaxIdle(2), MaxActive(1))
	require.Error(t, err)
}

func TestClose(t *testing.T) {
	p, nativePool, err := newPool()
	require.NoError(t, err)
	p.Close()

	options := nativePool.opt
	require.EqualValues(t, 0, nativePool.index)
	require.EqualValues(t, 0, nativePool.ref)
	require.EqualValues(t, 0, nativePool.current)
	require.EqualValues(t, true, nativePool.conns[0] == nil)
	require.EqualValues(t, true, nativePool.conns[options.maxIdle-1] == nil)
}

func TestReset(t *testing.T) {
	p, nativePool, err := newPool()
	require.NoError(t, err)
	defer p.Close()

	nativePool.delete(0)
	require.EqualValues(t, true, nativePool.conns[0] == nil)
	nativePool.delete(nativePool.opt.maxIdle + 1)
	require.EqualValues(t, true, nativePool.conns[nativePool.opt.maxIdle+1] == nil)
}

func TestBasicGet(t *testing.T) {
	p, nativePool, err := newPool()
	require.NoError(t, err)
	defer p.Close()

	conn, err := p.Get()
	require.NoError(t, err)
	require.EqualValues(t, true, conn.Value() != nil)

	require.EqualValues(t, 1, nativePool.index)
	require.EqualValues(t, 1, nativePool.ref)

	conn.Close()

	require.EqualValues(t, 1, nativePool.index)
	require.EqualValues(t, 0, nativePool.ref)
}

func TestAfterCloseGRPCChannel(t *testing.T) {
	p, _, err := newPool()
	require.NoError(t, err)
	defer p.Close()

	conn, err := p.Get()
	require.NoError(t, err)
	require.EqualValues(t, true, conn.Value() != nil)

	conn.Value().Close()
	t.Logf("after close GRPCChannel, the channel's state is :%s", conn.Value().GetState())
	time.Sleep(5 * time.Second)
	t.Logf("after close GRPCChannel 5s, the channel's state is :%s", conn.Value().GetState())
	require.EqualValues(t, true, conn.Value() != nil)
	require.EqualValues(t, true, conn.Value().GetState() == connectivity.Shutdown)
}

func TestGetAfterClose(t *testing.T) {
	p, _, err := newPool()
	require.NoError(t, err)
	p.Close()

	_, err = p.Get()
	require.EqualError(t, err, "pool is closed")
}

func TestBasicGet2(t *testing.T) {
	opts := []Option{
		Dial(DialTest),
		MaxIdle(1),
		MaxActive(2),
		MaxConcurrentStreams(2),
		Reuse(true),
	}

	p, nativePool, err := newPool(opts...)
	require.NoError(t, err)
	defer p.Close()

	conn1, err := p.Get()
	require.NoError(t, err)
	defer conn1.Close()

	conn2, err := p.Get()
	require.NoError(t, err)
	defer conn2.Close()

	require.EqualValues(t, 2, nativePool.index)
	require.EqualValues(t, 2, nativePool.ref)
	require.EqualValues(t, 1, nativePool.current)

	// create new connections push back to pool
	conn3, err := p.Get()
	require.NoError(t, err)
	defer conn3.Close()

	require.EqualValues(t, 3, nativePool.index)
	require.EqualValues(t, 3, nativePool.ref)
	require.EqualValues(t, 2, nativePool.current)

	conn4, err := p.Get()
	require.NoError(t, err)
	defer conn4.Close()

	// reuse exists connections
	conn5, err := p.Get()
	require.NoError(t, err)
	defer conn5.Close()

	nativeConn := conn5.(*conn)
	require.EqualValues(t, false, nativeConn.once)
}

func TestBasicGet3(t *testing.T) {
	opts := []Option{
		Dial(DialTest),
		MaxIdle(1),
		MaxActive(1),
		MaxConcurrentStreams(1),
		Reuse(false),
	}

	p, _, err := newPool(opts...)
	require.NoError(t, err)
	defer p.Close()

	conn1, err := p.Get()
	require.NoError(t, err)
	defer conn1.Close()

	// create new connections doesn't push back to pool
	conn2, err := p.Get()
	require.NoError(t, err)
	defer conn2.Close()

	nativeConn := conn2.(*conn)
	require.EqualValues(t, true, nativeConn.once)
}

func TestConcurrentGet(t *testing.T) {
	opts := []Option{
		Dial(DialTest),
		MaxIdle(8),
		MaxActive(64),
		MaxConcurrentStreams(2),
		Reuse(false),
	}

	p, nativePool, err := newPool(opts...)
	require.NoError(t, err)
	defer p.Close()

	var wg sync.WaitGroup
	wg.Add(500)

	for i := 0; i < 500; i++ {
		go func(i int) {
			conn, err := p.Get()
			require.NoError(t, err)
			require.EqualValues(t, true, conn != nil)
			conn.Close()
			wg.Done()
			t.Logf("goroutine: %v, index: %v, ref: %v, current: %v", i,
				atomic.LoadUint32(&nativePool.index),
				atomic.LoadInt32(&nativePool.ref),
				atomic.LoadInt32(&nativePool.current))
		}(i)
	}
	wg.Wait()

	require.EqualValues(t, 0, nativePool.ref)
	require.EqualValues(t, nativePool.opt.maxIdle, nativePool.current)
	require.EqualValues(t, true, nativePool.conns[0] != nil)
	require.EqualValues(t, true, nativePool.conns[nativePool.opt.maxIdle] == nil)
}

var size = 4 * 1024 * 1024

func BenchmarkPoolRPC(b *testing.B) {
	p, err := New(*endpoint)
	if err != nil {
		b.Fatalf("failed to new pool: %v", err)
	}
	defer p.Close()

	testFunc := func() {
		conn, err := p.Get()
		if err != nil {
			b.Fatalf("failed to get conn: %v", err)
		}
		defer conn.Close()

		client := pb.NewEchoClient(conn.Value())
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		data := make([]byte, size)
		_, err = client.Say(ctx, &pb.EchoRequest{Message: data})
		if err != nil {
			b.Fatalf("unexpected error from Say: %v", err)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(tpb *testing.PB) {
		for tpb.Next() {
			testFunc()
		}
	})
}

func BenchmarkSingleRPC(b *testing.B) {
	testFunc := func() {
		cc, err := DftDial(*endpoint)
		if err != nil {
			b.Fatalf("failed to create grpc conn: %v", err)
		}
		defer cc.Close()

		client := pb.NewEchoClient(cc)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		data := make([]byte, size)
		_, err = client.Say(ctx, &pb.EchoRequest{Message: data})
		if err != nil {
			b.Fatalf("unexpected error from Say: %v", err)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(tpb *testing.PB) {
		for tpb.Next() {
			testFunc()
		}
	})
}

func BenchmarkOnlyOneRPC(b *testing.B) {
	cc, err := DftDial(*endpoint)
	if err != nil {
		b.Logf("dial to grpc server fail %v", err)
	}
	defer cc.Close()

	testFunc := func() {
		client := pb.NewEchoClient(cc)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		data := make([]byte, size)
		_, err := client.Say(ctx, &pb.EchoRequest{Message: data})
		if err != nil {
			b.Fatalf("unexpected error from Say: %v", err)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(tpb *testing.PB) {
		for tpb.Next() {
			testFunc()
		}
	})
}
