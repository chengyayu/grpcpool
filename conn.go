package grpcpool

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// Conn single grpc connection interface
type Conn interface {
	// Value return the actual grpc connection type *grpc.ClientConn.
	Value() grpc.ClientConnInterface

	// Close decrease the reference of grpc connection, instead of close it.
	// if the pool is full, just close it.
	Close() error

	// GetState return the state of grpc connection
	GetState() connectivity.State
}

// Conn is wrapped grpc.ClientConn. to provide close and value method.
type conn struct {
	cc   *grpc.ClientConn
	pool *pool
	once bool
}

// Value see Conn interface.
func (c *conn) Value() grpc.ClientConnInterface {
	return c.cc
}

// Close see Conn interface.
func (c *conn) Close() error {
	c.pool.decrRef()
	if c.once {
		return c.reset()
	}
	return nil
}

// GetState see Conn interface.
func (c *conn) GetState() connectivity.State {
	return c.cc.GetState()
}

// 重置连接，让它等待垃圾回收
func (c *conn) reset() error {
	cc := c.cc
	c.cc = nil
	c.once = false
	if cc != nil {
		return cc.Close()
	}
	return nil
}
