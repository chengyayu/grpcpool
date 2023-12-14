package grpcpool

import (
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"sync"
)

// Pool interface describes a pool implementation.
// An ideal pool is thread-safe and easy to use.
type Pool interface {
	// Get returns a new connection from the pool. Closing the connections puts
	// it back to the Pool. Closing it when the pool is destroyed or full will
	// be counted as an error. we guarantee the conn.Value() isn't nil when conn isn't nil.
	Get() (Conn, error)

	// Close closes the pool and all its connections. After Close() the pool is
	// no longer usable. You can't make concurrent calls Close and Get method.
	// It will be cause panic.
	Close() error

	// Status returns the current status of the pool.
	Status() string
}

type pool struct {
	// atomic, used to get connection random.
	index uint32

	// control the atomic var current's concurrent read write.
	sync.RWMutex

	// atomic, the current physical connection of pool.
	current int32

	// atomic, the using logic connection of pool
	// logic connection = physical connection * maxConcurrentStreams
	ref int32

	// pool options
	opt options

	// all of created physical connections.
	conns []*conn

	// the server address is to create connection.
	address string

	// closed set true when Close is called.
	closed int32
}

// New return a connection pool.
func New(address string, opts ...Option) (Pool, error) {
	o := options{
		dial:                 DefaultDial,
		maxIdle:              8,
		maxActive:            64,
		maxConcurrentStreams: 64,
		reuse:                true,
	}

	for _, opt := range opts {
		opt(&o)
	}

	if address == "" {
		return nil, errors.New("invalid address settings")
	}
	if o.dial == nil {
		return nil, errors.New("invalid dial settings")
	}
	if o.maxIdle <= 0 || o.maxActive <= 0 || o.maxIdle > o.maxActive {
		return nil, errors.New("invalid maximum settings")
	}
	if o.maxConcurrentStreams <= 0 {
		return nil, errors.New("invalid maxConcurrentStreams settings")
	}

	p := &pool{
		current: int32(o.maxIdle),
		opt:     o,
		conns:   make([]*conn, o.maxActive),
		address: address,
	}

	for i := 0; i < p.opt.maxIdle; i++ {
		c, err := p.opt.dial(address)
		if err != nil {
			p.Close()
			return nil, fmt.Errorf("dial is not able to fill the pool: %s", err)
		}
		p.conns[i] = p.wrapConn(c, false)
	}
	log.Printf("new pool success: %v\n", p.Status())

	return p, nil
}

func (p *pool) Get() (Conn, error) {
	//TODO implement me
	panic("implement me")
}

func (p *pool) Close() error {
	//TODO implement me
	panic("implement me")
}

func (p *pool) Status() string {
	//TODO implement me
	panic("implement me")
}

func (p *pool) wrapConn(cc *grpc.ClientConn, once bool) *conn {
	return &conn{
		cc:   cc,
		pool: p,
		once: once,
	}
}
