package grpcpool

import (
	"context"
	"flag"
	"github.com/chengyayu/grpcpool/example/pb"
	"github.com/stretchr/testify/require"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var endpoint = flag.String("endpoint", "127.0.0.1:40000", "grpc server endpoint")

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
		cc, err := DialTest(*endpoint)
		if err != nil {
			b.Fatalf("failed to create grpc conn: %v", err)
		}

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
