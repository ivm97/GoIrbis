package irbis

import "context"

// Service is the stable surface for application code and tests.
// *Client implements it with a one-shot session per call.
type Service interface {
	Do(ctx context.Context, fn func(*Connection) error) error
	Search(ctx context.Context, expression string) ([]int, error)
	SearchCount(ctx context.Context, expression string) (int, error)
	ReadRecord(ctx context.Context, mfn int) (*MarcRecord, error)
	WriteRecord(ctx context.Context, record *MarcRecord) (int, error)
	FormatMfn(ctx context.Context, format string, mfn int) (string, error)
	NoOp(ctx context.Context) error
}

// Compile-time check: Client satisfies Service.
var _ Service = (*Client)(nil)
