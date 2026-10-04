package irbis

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"
)

// scriptSocket returns scripted IRBIS replies; each TalkToServer consumes one.
type scriptSocket struct {
	conn      *Connection
	responses [][]byte
	calls     int
	lastCtx   context.Context
}

func (s *scriptSocket) TalkToServer(ctx context.Context, _ *ClientQuery) (*ServerResponse, error) {
	s.lastCtx = ctx
	if err := ctx.Err(); err != nil {
		return nil, WrapError(ErrCodeNetwork, err)
	}
	if s.calls >= len(s.responses) {
		return nil, fmt.Errorf("unexpected query #%d", s.calls+1)
	}
	response := ParseServerResponse(s.responses[s.calls])
	s.calls++
	response.connection = s.conn
	return response, nil
}

// reply builds a minimal ANSI server response body.
// lines: command, clientId, queryId, answerSize, version, 5 stubs, then payload lines.
func reply(command string, returnCode int, rest ...string) []byte {
	var buf bytes.Buffer
	write := func(s string) {
		buf.WriteString(s)
		buf.WriteByte('\r')
		buf.WriteByte('\n')
	}
	write(command)
	write("1")
	write("1")
	write("0")
	write("64")
	for i := 0; i < 5; i++ {
		write("")
	}
	write(fmt.Sprintf("%d", returnCode))
	for _, line := range rest {
		write(line)
	}
	return buf.Bytes()
}

func TestConnectContext_Success(t *testing.T) {
	conn := NewConnection()
	sock := &scriptSocket{
		conn: conn,
		responses: [][]byte{
			reply("A", 0, "5", "[Main]", "DBNNAMECAT=1"),
		},
	}
	conn.socket = sock

	if err := conn.ConnectContext(context.Background()); err != nil {
		t.Fatalf("ConnectContext: %v", err)
	}
	if !conn.Connected {
		t.Fatal("expected Connected=true")
	}
	if conn.Interval != 5 {
		t.Fatalf("Interval=%d, want 5", conn.Interval)
	}
	if got := conn.Ini.GetValue("Main", "DBNNAMECAT", ""); got != "1" {
		t.Fatalf("INI DBNNAMECAT=%q, want 1", got)
	}
}

func TestConnectContext_Canceled(t *testing.T) {
	conn := NewConnection()
	conn.socket = &scriptSocket{conn: conn}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := conn.ConnectContext(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
	if CodeOf(err) != ErrCodeNetwork {
		t.Fatalf("CodeOf=%d, want %d (%v)", CodeOf(err), ErrCodeNetwork, err)
	}
}

func TestExecuteContext_IncrementsQueryId(t *testing.T) {
	conn := NewConnection()
	conn.Connected = true
	conn.ClientId = 123456
	conn.QueryId = 1
	sock := &scriptSocket{
		conn: conn,
		responses: [][]byte{
			reply("N", 0),
			reply("N", 0),
		},
	}
	conn.socket = sock

	q1 := NewClientQuery(conn, "N")
	if conn.QueryId != 2 {
		t.Fatalf("QueryId after first NewClientQuery=%d, want 2", conn.QueryId)
	}
	if _, err := conn.ExecuteContext(context.Background(), q1); err != nil {
		t.Fatalf("ExecuteContext #1: %v", err)
	}

	q2 := NewClientQuery(conn, "N")
	if conn.QueryId != 3 {
		t.Fatalf("QueryId after second NewClientQuery=%d, want 3", conn.QueryId)
	}
	if _, err := conn.ExecuteContext(context.Background(), q2); err != nil {
		t.Fatalf("ExecuteContext #2: %v", err)
	}
}

func TestClient_Search_OneShotSession(t *testing.T) {
	conn := NewConnection(Config{
		Username: "librarian",
		Password: "secret",
	})
	sock := &scriptSocket{
		conn: conn,
		responses: [][]byte{
			reply("A", 0, "5"),          // login
			reply("K", 0, "2", "10", "11"), // search: count + mfns
			reply("B", 0),               // logout
		},
	}
	conn.socket = sock

	if err := conn.ConnectContext(context.Background()); err != nil {
		t.Fatalf("connect: %v", err)
	}
	found := conn.Search(`"A=TEST$"`)
	if len(found) != 2 || found[0] != 10 || found[1] != 11 {
		t.Fatalf("found=%v, want [10 11]", found)
	}
	if err := conn.DisconnectContext(context.Background()); err != nil {
		t.Fatalf("disconnect: %v", err)
	}
	if sock.calls != 3 {
		t.Fatalf("socket calls=%d, want 3 (A,K,B)", sock.calls)
	}
}

func TestClient_Do_UsesContextDeadline(t *testing.T) {
	conn := NewConnection()
	sock := &scriptSocket{conn: conn}
	conn.socket = sock

	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	time.Sleep(time.Millisecond)

	err := conn.ConnectContext(ctx)
	if err == nil {
		t.Fatal("expected deadline error")
	}
	if CodeOf(err) != ErrCodeNetwork {
		t.Fatalf("CodeOf=%d, want %d", CodeOf(err), ErrCodeNetwork)
	}
}

func TestCodeOf(t *testing.T) {
	err := NewError(-4444)
	if CodeOf(err) != -4444 {
		t.Fatalf("CodeOf=%d, want -4444", CodeOf(err))
	}
	if CodeOf(fmt.Errorf("plain")) != 0 {
		t.Fatal("plain error should yield 0")
	}
}
