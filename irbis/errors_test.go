package irbis

import "testing"

func TestDescribeError_Success(t *testing.T) {
	if got := DescribeError(0); got != "no error" {
		t.Fatalf("DescribeError(0)=%q, want %q", got, "no error")
	}
	if got := DescribeError(1); got != "no error" {
		t.Fatalf("DescribeError(1)=%q, want %q", got, "no error")
	}
}

func TestDescribeError_KnownCodes(t *testing.T) {
	cases := map[int]string{
		-3333:   "unregistered client (not in the client list)",
		-3337:   "client is already registered",
		-4444:   "wrong password",
		-5555:   "file does not exist",
		-602:    "record is locked for edit",
		-608:    "record version conflict",
		-100000: "network error: failed to connect to server",
	}
	for code, want := range cases {
		if got := DescribeError(code); got != want {
			t.Fatalf("DescribeError(%d)=%q, want %q", code, got, want)
		}
	}
}

func TestDescribeError_UnknownCodeIncludesNumber(t *testing.T) {
	got := DescribeError(-424242)
	want := "unknown error (-424242)"
	if got != want {
		t.Fatalf("DescribeError(-424242)=%q, want %q", got, want)
	}
}

func TestRawRecordEncode_UsesDelimiter(t *testing.T) {
	rec := NewRawRecord()
	rec.Mfn = 12
	rec.Status = 0
	rec.Version = 3
	rec.Fields = []string{"200#^aTitle"}

	got := rec.Encode(FullDelimiter)
	want := "12#0" + FullDelimiter + "0#3" + FullDelimiter + "200#^aTitle" + FullDelimiter
	if got != want {
		t.Fatalf("Encode()=%q, want %q", got, want)
	}
}
