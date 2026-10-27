package irbis

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// captureSocket records the last query body and returns scripted replies.
type captureSocket struct {
	conn      *Connection
	responses [][]byte
	calls     int
	lastBody  []byte
}

func (s *captureSocket) TalkToServer(_ context.Context, query *ClientQuery) (*ServerResponse, error) {
	s.lastBody = append([]byte(nil), query.buf...)
	if s.calls >= len(s.responses) {
		return nil, errUnexpectedQuery(s.calls + 1)
	}
	response := ParseServerResponse(s.responses[s.calls])
	s.calls++
	response.connection = s.conn
	return response, nil
}

type unexpectedQueryError int

func errUnexpectedQuery(n int) error { return unexpectedQueryError(n) }
func (e unexpectedQueryError) Error() string {
	return "unexpected query"
}

func TestSearchFull_ReturnsTotal(t *testing.T) {
	conn := NewConnection()
	conn.Connected = true
	conn.Database = "IBIS"
	sock := &captureSocket{
		conn: conn,
		responses: [][]byte{
			reply("K", 0, "42", "10#brief", "11#brief"),
		},
	}
	conn.socket = sock

	params := NewSearchParameters()
	params.Expression = `"A=TEST$"`
	params.NumberOfRecords = 2
	params.FirstRecord = 1
	params.Format = "@brief"

	out := conn.SearchFull(params)
	if out.Total != 42 {
		t.Fatalf("Total=%d, want 42", out.Total)
	}
	if len(out.Found) != 2 || out.Found[0].Mfn != 10 || out.Found[1].Mfn != 11 {
		t.Fatalf("Found=%+v", out.Found)
	}
}

func TestGlobalCorrection_RangeMode(t *testing.T) {
	conn := NewConnection()
	conn.Connected = true
	conn.Database = "IBIS"
	sock := &captureSocket{
		conn:      conn,
		responses: [][]byte{reply("5", 0, "ok")},
	}
	conn.socket = sock

	settings := &GblSettings{
		Database:         "IBIS",
		Actualize:        true,
		SearchExpression: "*",
		FirstRecord:      1,
		NumberOfRecords:  0,
		MinMfn:           1,
		MaxMfn:           100,
		Statements: []GblStatement{
			{Command: ADD_FIELD, Parameter1: "907", Parameter2: "*", Format1: "'x'", Format2: ""},
		},
	}
	_ = conn.GlobalCorrection(settings)

	body := string(sock.lastBody)
	// Range mode must send 0, min, max — not expand MFNs.
	if !strings.Contains(body, "\n0\n1\n100\n") {
		t.Fatalf("expected range protocol 0/1/100 in body:\n%s", body)
	}
	if bytes.Contains(sock.lastBody, []byte("\n50\n")) {
		t.Fatal("must not expand MFN range into individual numbers")
	}
}

func TestGlobalCorrection_MfnList(t *testing.T) {
	conn := NewConnection()
	conn.Connected = true
	sock := &captureSocket{
		conn:      conn,
		responses: [][]byte{reply("5", 0)},
	}
	conn.socket = sock

	settings := &GblSettings{
		MfnList:    []int{5, 7, 9},
		Statements: []GblStatement{{Command: DELETE_FIELD, Parameter1: "999", Parameter2: "*"}},
	}
	_ = conn.GlobalCorrection(settings)

	body := string(sock.lastBody)
	if !strings.Contains(body, "\n3\n5\n7\n9\n") {
		t.Fatalf("expected MFN list protocol:\n%s", body)
	}
}
