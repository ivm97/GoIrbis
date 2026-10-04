package irbis

import "time"

const (
	// DefaultDialTimeout is used when Config.DialTimeout is zero.
	DefaultDialTimeout = 10 * time.Second

	// DefaultHost is the IRBIS server address used by empty Config.
	DefaultHost = "127.0.0.1"
	// DefaultPort is the standard IRBIS64 port.
	DefaultPort = 6666
	// DefaultDatabase is the catalog name used by empty Config.
	DefaultDatabase = "IBIS"
	// DefaultWorkstation is the cataloger ARM code.
	DefaultWorkstation = "C"
)

// Config holds connection settings for Client and Connection.
//
// Timeouts are client-side only: when they fire, the TCP wait is aborted.
// The IRBIS server may keep working on the command; it is not stopped remotely.
//
// DialTimeout: max time to establish TCP. Zero becomes DefaultDialTimeout.
// IOTimeout: max time for one command (write+read) when context has no deadline.
// Zero means no transport IO deadline — only context cancellation applies.
// That is the recommended default for slow IRBIS searches and formats.
type Config struct {
	Host        string
	Port        int
	Username    string
	Password    string
	Database    string
	Workstation string
	DialTimeout time.Duration
	IOTimeout   time.Duration
}

// withDefaults fills empty address fields and DialTimeout.
// IOTimeout is left as-is (zero = unlimited at transport level).
func (c Config) withDefaults() Config {
	if c.Host == "" {
		c.Host = DefaultHost
	}
	if c.Port == 0 {
		c.Port = DefaultPort
	}
	if c.Database == "" {
		c.Database = DefaultDatabase
	}
	if c.Workstation == "" {
		c.Workstation = DefaultWorkstation
	}
	if c.DialTimeout == 0 {
		c.DialTimeout = DefaultDialTimeout
	}
	return c
}
