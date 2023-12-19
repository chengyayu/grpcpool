package grpcpool

type IConnStateHandler interface {
	HandleIdle(c *conn) (Conn, error)
	HandleConnecting(c *conn) (Conn, error)
	HandleReady(c *conn) (Conn, error)
	HandleTransientFailure(c *conn) (Conn, error)
	HandleShutDown(c *conn) (Conn, error)
}
