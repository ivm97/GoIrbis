package irbis

import (
	"bytes"
	"testing"
)

func TestCP1251_RoundTripCyrillic(t *testing.T) {
	in := "Пушкин A=TEST"
	encoded := ToAnsi(in)
	decoded := FromAnsi(encoded)
	if decoded != in {
		t.Fatalf("round-trip: got %q, want %q", decoded, in)
	}
}

func TestCP1251_ASCIIFastPath(t *testing.T) {
	in := "IBIS"
	got := ToAnsi(in)
	if !bytes.Equal(got, []byte(in)) {
		t.Fatalf("ASCII encode=%v", got)
	}
	if FromAnsi(got) != in {
		t.Fatalf("ASCII decode=%q", FromAnsi(got))
	}
}

func TestClientQuery_EncodePacketSingleWrite(t *testing.T) {
	conn := NewConnection()
	conn.ClientId = 100001
	conn.QueryId = 1
	conn.Username = "user"
	conn.Password = "pass"
	conn.Workstation = "C"

	q := NewClientQuery(conn, "K")
	q.AddAnsi("IBIS").NewLine()
	q.AddUtf(`"A=TEST$"`).NewLine()

	packet := q.EncodePacket()
	if len(packet) == 0 || packet[0] < '0' || packet[0] > '9' {
		t.Fatalf("packet must start with decimal length, got %q", packet[:min(16, len(packet))])
	}
	if bytes.Count(packet, []byte{'\n'}) < 8 {
		t.Fatalf("expected multi-line request, newlines=%d", bytes.Count(packet, []byte{'\n'}))
	}
	if conn.QueryId != 2 {
		t.Fatalf("QueryId=%d, want 2", conn.QueryId)
	}
}

func TestCheckReturnCode_AllowedClearsError(t *testing.T) {
	conn := NewConnection()
	raw := reply("H", CodeTermNotFound)
	response := ParseServerResponse(raw)
	response.connection = conn

	if !response.CheckReturnCode(CodeTermNotFound, CodeLastTerm, CodeFirstTerm) {
		t.Fatal("allowed code must succeed")
	}
	if conn.Err() != nil {
		t.Fatalf("allowed code must clear lastErr, got %v", conn.Err())
	}
}

func BenchmarkToAnsi_Cyrillic(b *testing.B) {
	text := "Александр Сергеевич Пушкин — Евгений Онегин"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ToAnsi(text)
	}
}

func BenchmarkToAnsi_ASCII(b *testing.B) {
	text := "3.IBIS.brief.pft"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ToAnsi(text)
	}
}

func BenchmarkClientQuery_EncodePacket(b *testing.B) {
	conn := NewConnection()
	conn.ClientId = 123456
	conn.Username = "librarian"
	conn.Password = "secret"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		conn.QueryId = 1
		q := NewClientQuery(conn, "K")
		q.AddAnsi("IBIS").NewLine()
		q.AddUtf(`"A=ПУШКИН$"`).NewLine()
		q.Add(0).NewLine()
		q.Add(1).NewLine()
		_ = q.EncodePacket()
	}
}
