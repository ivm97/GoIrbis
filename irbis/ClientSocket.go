package irbis

import (
	"context"
	"net"
	"strconv"
	"time"
)

// ClientSocket sends one client query and reads one server response.
// IRBIS uses a fresh TCP connection per command.
type ClientSocket interface {
	TalkToServer(ctx context.Context, query *ClientQuery) (*ServerResponse, error)
}

// Tcp4ClientSocket is the default TCP transport.
type Tcp4ClientSocket struct {
	connection *Connection
}

// NewTcp4ClientSocket binds a socket implementation to a connection.
func NewTcp4ClientSocket(connection *Connection) *Tcp4ClientSocket {
	return &Tcp4ClientSocket{connection: connection}
}

func (client *Tcp4ClientSocket) TalkToServer(ctx context.Context, query *ClientQuery) (*ServerResponse, error) {
	connection := client.connection
	if ctx == nil {
		ctx = context.Background()
	}

	// Aborting here only closes the client TCP wait. IRBIS may still finish the command.
	address := net.JoinHostPort(connection.Host, strconv.Itoa(connection.Port))
	dialer := net.Dialer{Timeout: connection.dialTimeout()}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, WrapError(ErrCodeNetwork, err)
	}
	defer func() { _ = conn.Close() }()

	if err := applyIODeadline(ctx, conn, connection.ioTimeout()); err != nil {
		return nil, WrapError(ErrCodeNetwork, err)
	}

	packet := query.EncodePacket()
	if _, err := conn.Write(packet); err != nil {
		return nil, WrapError(ErrCodeNetwork, err)
	}

	response, err := ReadServerResponse(conn)
	if err != nil {
		return nil, WrapError(ErrCodeNetwork, err)
	}
	return response, nil
}

// applyIODeadline sets a socket deadline from context and/or IOTimeout.
// IOTimeout <= 0 means no transport limit beyond context.
func applyIODeadline(ctx context.Context, conn net.Conn, ioTimeout time.Duration) error {
	deadline, hasCtxDeadline := ctx.Deadline()
	if ioTimeout > 0 {
		alt := time.Now().Add(ioTimeout)
		if !hasCtxDeadline || alt.Before(deadline) {
			deadline = alt
			hasCtxDeadline = true
		}
	}
	if !hasCtxDeadline {
		return nil
	}
	return conn.SetDeadline(deadline)
}
