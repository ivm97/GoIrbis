package irbis

import (
	"context"
	"time"
)

// Client is a stateless IRBIS configuration.
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

// NewClient builds a client from Config.
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
		return conn.Err()
	})
	return found, err
}

// SearchCount returns the number of records matching expression.
func (c *Client) SearchCount(ctx context.Context, expression string) (int, error) {
	var count int
	err := c.Do(ctx, func(conn *Connection) error {
		count = conn.SearchCount(expression)
		return conn.Err()
	})
	return count, err
}

// SearchAll returns all matching MFNs, paging past the 32k server limit.
func (c *Client) SearchAll(ctx context.Context, expression string) ([]int, error) {
	var found []int
	err := c.Do(ctx, func(conn *Connection) error {
		found = conn.SearchAll(expression)
		return conn.Err()
	})
	return found, err
}

// SearchRead searches and loads records in one session.
func (c *Client) SearchRead(ctx context.Context, expression string, limit int) ([]MarcRecord, error) {
	var records []MarcRecord
	err := c.Do(ctx, func(conn *Connection) error {
		records = conn.SearchRead(expression, limit)
		return conn.Err()
	})
	return records, err
}

// ReadRecord loads one record by MFN.
func (c *Client) ReadRecord(ctx context.Context, mfn int) (*MarcRecord, error) {
	var record *MarcRecord
	err := c.Do(ctx, func(conn *Connection) error {
		record = conn.ReadRecord(mfn)
		if record == nil {
			if err := conn.Err(); err != nil {
				return err
			}
			return NewError(CodeNetwork)
		}
		return nil
	})
	return record, err
}

// ReadRecords loads several records in one session.
func (c *Client) ReadRecords(ctx context.Context, mfns []int) ([]MarcRecord, error) {
	var records []MarcRecord
	err := c.Do(ctx, func(conn *Connection) error {
		records = conn.ReadRecords(mfns)
		return conn.Err()
	})
	return records, err
}

// WriteRecord saves a record and returns the server max MFN.
func (c *Client) WriteRecord(ctx context.Context, record *MarcRecord) (int, error) {
	var maxMfn int
	err := c.Do(ctx, func(conn *Connection) error {
		maxMfn = conn.WriteRecord(record)
		return conn.Err()
	})
	return maxMfn, err
}

// FormatMfn formats a stored record on the server.
func (c *Client) FormatMfn(ctx context.Context, format string, mfn int) (string, error) {
	var text string
	err := c.Do(ctx, func(conn *Connection) error {
		text = conn.FormatMfn(format, mfn)
		return conn.Err()
	})
	return text, err
}

// GetMaxMfn returns the database max MFN.
func (c *Client) GetMaxMfn(ctx context.Context, database string) (int, error) {
	var maxMfn int
	err := c.Do(ctx, func(conn *Connection) error {
		maxMfn = conn.GetMaxMfn(database)
		return conn.Err()
	})
	return maxMfn, err
}

// ReadTerms reads dictionary terms starting from startTerm.
func (c *Client) ReadTerms(ctx context.Context, startTerm string, number int) ([]TermInfo, error) {
	var terms []TermInfo
	err := c.Do(ctx, func(conn *Connection) error {
		terms = conn.ReadTerms(startTerm, number)
		return conn.Err()
	})
	return terms, err
}

// ReadTextFile reads a server-side text file.
func (c *Client) ReadTextFile(ctx context.Context, specification string) (string, error) {
	var text string
	err := c.Do(ctx, func(conn *Connection) error {
		text = conn.ReadTextFile(specification)
		return conn.Err()
	})
	return text, err
}

// NoOp confirms the session is alive (one-shot login + NoOp + logout).
func (c *Client) NoOp(ctx context.Context) error {
	return c.Do(ctx, func(conn *Connection) error {
		if !conn.NoOp() {
			if err := conn.Err(); err != nil {
				return err
			}
			return NewError(CodeNetwork)
		}
		return nil
	})
}
