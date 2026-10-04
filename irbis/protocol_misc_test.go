package irbis

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRawRecord_DecodeEncode(t *testing.T) {
	rec := NewRawRecord()
	rec.Decode([]string{
		"12#0",
		"0#3",
		"200#^aTitle",
		"700#^aAuthor",
		"",
	})
	if rec.Mfn != 12 || rec.Version != 3 || len(rec.Fields) != 2 {
		t.Fatalf("decoded=%+v", rec)
	}
	got := rec.Encode(FullDelimiter)
	want := "12#0" + FullDelimiter + "0#3" + FullDelimiter + "200#^aTitle" + FullDelimiter + "700#^aAuthor" + FullDelimiter
	if got != want {
		t.Fatalf("Encode()=%q", got)
	}
}

func TestIniSection_Remove(t *testing.T) {
	sec := &IniSection{Name: "Main", Lines: []IniLine{
		{Key: "A", Value: "1"},
		{Key: "B", Value: "2"},
	}}
	sec.Remove("a") // case-insensitive
	if len(sec.Lines) != 1 || sec.Lines[0].Key != "B" {
		t.Fatalf("after Remove: %+v", sec.Lines)
	}
	sec.SetValue("B", "")
	if len(sec.Lines) != 0 {
		t.Fatalf("SetValue empty should remove, got %+v", sec.Lines)
	}
}

func TestSplitLines_CRLF(t *testing.T) {
	got := SplitLines("a\r\nb\rc")
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("SplitLines=%q", got)
	}
}

func TestTreeFile_ParseFixture(t *testing.T) {
	path := filepath.Join("..", "data", "test1.tre")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skip(err)
	}
	tree := new(TreeFile)
	if err := tree.Parse(SplitLines(string(raw))); err != nil {
		t.Fatal(err)
	}
	if len(tree.Roots) != 4 {
		t.Fatalf("roots=%d, want 4", len(tree.Roots))
	}
	if tree.Roots[0].Value != "1 - First" {
		t.Fatalf("root0=%q", tree.Roots[0].Value)
	}
	second := tree.Roots[1]
	if len(second.Children) != 3 {
		t.Fatalf("second children=%d, want 3", len(second.Children))
	}
	if len(second.Children[1].Children) != 1 {
		t.Fatalf("2.2 children=%d, want 1", len(second.Children[1].Children))
	}
}

func TestTreeFile_ParseInvalidIndent(t *testing.T) {
	tree := new(TreeFile)
	err := tree.Parse([]string{"\tbad"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConstantAliases(t *testing.T) {
	if StatusLocked != LOCKED_RECORD || FormatBrief != BRIEF_FORMAT {
		t.Fatal("alias mismatch")
	}
}
