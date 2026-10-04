package irbis

import (
	"context"
	"time"
)

// Client is a stateless IRBIS configuration that implements Service.
// Each method opens a fresh logical session: login → work → logout.
// Safe for concurrent use: every call uses its own Connection.
type Client struct {
	Host        string
	Port        int
	Username    string
	Password    string
	Database    string
	Workstation string
	DialTimeout time.Duration
	IOTimeout   time.Duration
}

// NewClient builds a Service implementation from Config.
// Omit Config for package defaults (IOTimeout=0: wait until context cancels).
func NewClient(cfg ...Config) *Client {
	var c Config
	if len(cfg) > 0 {
		c = cfg[0]
	}
	c = c.withDefaults()
	return &Client{
		Host:        c.Host,
		Port:        c.Port,
		Username:    c.Username,
		Password:    c.Password,
		Database:    c.Database,
		Workstation: c.Workstation,
		DialTimeout: c.DialTimeout,
		IOTimeout:   c.IOTimeout,
	}
}

func (c *Client) newConnection() *Connection {
	return NewConnection(Config{
		Host:        c.Host,
		Port:        c.Port,
		Username:    c.Username,
		Password:    c.Password,
		Database:    c.Database,
		Workstation: c.Workstation,
		DialTimeout: c.DialTimeout,
		IOTimeout:   c.IOTimeout,
	})
}

// Do runs fn inside a short-lived session bound to ctx.
func (c *Client) Do(ctx context.Context, fn func(*Connection) error) error {
	if ctx == nil {
		ctx = context.Background()
	}

	conn := c.newConnection()
	if err := conn.ConnectContext(ctx); err != nil {
		return err
	}

	// Logout should not depend on a possibly already-expired request context.
	defer func() {
		_ = conn.DisconnectContext(context.Background())
	}()

	return fn(conn)
}

// Search runs a search in a one-shot session.
func (c *Client) Search(ctx context.Context, expression string) ([]int, error) {
	var found []int
	err := c.Do(ctx, func(conn *Connection) error {
		found = conn.Search(expression)
		if conn.LastError < 0 {
			return conn.Err()
		}
		return nil
	})
	return found, err
}

// SearchCount returns the number of records matching expression.
func (c *Client) SearchCount(ctx context.Context, expression string) (int, error) {
	var count int
	err := c.Do(ctx, func(conn *Connection) error {
		count = conn.SearchCount(expression)
		if conn.LastError < 0 {
			return conn.Err()
		}
		return nil
	})
	return count, err
}

// ReadRecord loads one record by MFN.
func (c *Client) ReadRecord(ctx context.Context, mfn int) (*MarcRecord, error) {
	var record *MarcRecord
	err := c.Do(ctx, func(conn *Connection) error {
		record = conn.ReadRecord(mfn)
		if record == nil {
			if conn.LastError < 0 {
				return conn.Err()
			}
			return NewError(ErrCodeNetwork)
		}
		return nil
	})
	return record, err
}

// WriteRecord saves a record and returns the server max MFN.
func (c *Client) WriteRecord(ctx context.Context, record *MarcRecord) (int, error) {
	var maxMfn int
	err := c.Do(ctx, func(conn *Connection) error {
		maxMfn = conn.WriteRecord(record)
		if conn.LastError < 0 {
			return conn.Err()
		}
		return nil
	})
	return maxMfn, err
}

// FormatMfn formats a stored record on the server.
func (c *Client) FormatMfn(ctx context.Context, format string, mfn int) (string, error) {
	var text string
	err := c.Do(ctx, func(conn *Connection) error {
		text = conn.FormatMfn(format, mfn)
		if conn.LastError < 0 {
			return conn.Err()
		}
		return nil
	})
	return text, err
}

// NoOp confirms the session is alive (one-shot login + NoOp + logout).
func (c *Client) NoOp(ctx context.Context) error {
	return c.Do(ctx, func(conn *Connection) error {
		if !conn.NoOp() {
			if conn.LastError < 0 {
				return conn.Err()
			}
			return NewError(ErrCodeNetwork)
		}
		return nil
	})
}
