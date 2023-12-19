package grpcpool

import (
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"math"
	"sync"
	"sync/atomic"
)

// ErrClosed is the error resulting if the pool is closed via pool.Close().
var ErrClosed = errors.New("pool is closed")

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
	Close()

	// Status returns the current status of the pool.
	Status() string
}

type pool struct {
	// atomic, used to get connection random.
	index uint32

	// atomic, the current physical connection of pool.
	current int32

	// atomic, the using logic connection of pool
	// logic connection = physical connection * options.maxConcurrentStreams
	ref int32

	// pool options
	opt options

	// all of created physical connections.
	conns []*conn

	// the server address is to create connection.
	address string

	// closed set true when Close is called.
	closed int32

	// control the atomic var current's concurrent read write.
	sync.RWMutex
}

func New(address string, opts ...Option) (Pool, error) {
	o := options{
		dial:                 DftDial,
		maxIdle:              DftMaxIdle,
		maxActive:            DftMaxActive,
		maxConcurrentStreams: DftMaxConcurrentStreams,
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
	//log.Printf("new pool success: %v\n", p.Status())

	return p, nil
}

func (p *pool) Get() (Conn, error) {
	nextRef := p.incrRef()
	p.RLock()
	current := atomic.LoadInt32(&p.current)
	p.RUnlock()
	if current == 0 {
		return nil, ErrClosed
	}

	// 现存逻辑连接数未被占满
	if nextRef <= current*int32(p.opt.maxConcurrentStreams) {
		next := atomic.AddUint32(&p.index, 1) % uint32(current)
		return p.conns[next], nil
	}

	// 物理连接数已达上限
	if current == int32(p.opt.maxActive) {
		// 开启了连接复用，从池中拿一个物理连接
		if p.opt.reuse {
			next := atomic.AddUint32(&p.index, 1) % uint32(current)
			return p.conns[next], nil
		}
		// 未开启连接复用，创建一次性物理连接
		c, err := p.opt.dial(p.address)
		return p.wrapConn(c, true), err
	}

	// 物理连接数未达上限，创建新的物理连接，放入池中
	p.Lock()
	current = atomic.LoadInt32(&p.current)
	if current < int32(p.opt.maxActive) && nextRef > current*int32(p.opt.maxConcurrentStreams) {
		// 2 times the incremental or the remain incremental
		increment := current
		if current+increment > int32(p.opt.maxActive) {
			increment = int32(p.opt.maxActive) - current
		}
		var i int32
		var err error
		for i = 0; i < increment; i++ {
			c, er := p.opt.dial(p.address)
			if er != nil {
				err = er
				break
			}
			p.delete(int(current + i))
			p.conns[current+i] = p.wrapConn(c, false)
		}
		current += i
		//log.Printf("grow pool: %d ---> %d, increment: %d, maxActive: %d\n", p.current, current, increment, p.opt.maxActive)
		atomic.StoreInt32(&p.current, current)
		if err != nil {
			p.Unlock()
			return nil, err
		}
	}
	p.Unlock()
	next := atomic.AddUint32(&p.index, 1) % uint32(current)
	return p.conns[next], nil
}

func (p *pool) Close() {
	atomic.StoreInt32(&p.closed, 1)
	atomic.StoreUint32(&p.index, 0)
	atomic.StoreInt32(&p.current, 0)
	atomic.StoreInt32(&p.ref, 0)
	p.deleteFrom(0)
	//log.Printf("close pool success: %v\n", p.Status())
}

func (p *pool) Status() string {
	return fmt.Sprintf("ptr: %p, address:%s, closed:%d, index:%d, current:%d, ref:%d. option:%v",
		p, p.address, p.closed, p.index, p.current, p.ref, p.opt)
}

func (p *pool) wrapConn(cc *grpc.ClientConn, once bool) *conn {
	return &conn{
		cc:   cc,
		pool: p,
		once: once,
	}
}

// 原子操作，引用计数（逻辑连接数）加一。
func (p *pool) incrRef() int32 {
	newRef := atomic.AddInt32(&p.ref, 1)
	if newRef == math.MaxInt32 {
		panic(fmt.Sprintf("overflow ref: %d", newRef))
	}
	return newRef
}

// 原子操作，引用计数（逻辑连接数）减一。
func (p *pool) decrRef() {
	newRef := atomic.AddInt32(&p.ref, -1)
	if newRef < 0 && atomic.LoadInt32(&p.closed) == 0 {
		panic(fmt.Sprintf("negative ref: %d", newRef))
	}
	// 无引用，当前物理连接数均为空闲连接，且超过了最大空闲连接数
	// 连接池缩容，将物理连接数减少至最大空闲连接数。
	if newRef == 0 && atomic.LoadInt32(&p.current) > int32(p.opt.maxIdle) {
		p.Lock()
		if atomic.LoadInt32(&p.ref) == 0 {
			//log.Printf("shrink pool: %d ---> %d, decrement: %d, maxActive: %d\n", p.current, p.opt.maxIdle, p.current-int32(p.opt.maxIdle), p.opt.maxActive)
			atomic.StoreInt32(&p.current, int32(p.opt.maxIdle))
			p.deleteFrom(p.opt.maxIdle)
		}
		p.Unlock()
	}
}

func (p *pool) deleteFrom(begin int) {
	for i := begin; i < p.opt.maxActive; i++ {
		p.delete(i)
	}
}

func (p *pool) delete(index int) {
	conn := p.conns[index]
	if conn == nil {
		return
	}
	_ = conn.reset()
	p.conns[index] = nil
}
