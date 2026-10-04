package irbis

import (
	"os"
	"path/filepath"
	"testing"
)

func dataFile(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "data", name)
	if _, err := os.Stat(path); err != nil {
		t.Skip(err)
	}
	return path
}

func TestFixture_IniFile(t *testing.T) {
	raw, err := os.ReadFile(dataFile(t, "inifile1.ini"))
	if err != nil {
		t.Fatal(err)
	}
	ini := NewIniFile()
	ini.Parse(SplitLines(string(raw)))
	if got := ini.GetValue("Main", "FirstParameter", ""); got != "1" {
		t.Fatalf("FirstParameter=%q", got)
	}
	if got := ini.GetValue("Private", "FourthParameter", ""); got != "4" {
		t.Fatalf("FourthParameter=%q", got)
	}
}

func TestFixture_ParFile(t *testing.T) {
	raw, err := os.ReadFile(dataFile(t, "ibis.par"))
	if err != nil {
		t.Fatal(err)
	}
	par := NewParFile("")
	par.Parse(SplitLines(string(raw)))
	if par.Mst == "" || par.Xrf == "" || par.Ifp == "" {
		t.Fatalf("par paths empty: %+v", par)
	}
}

func TestFixture_MenuFile(t *testing.T) {
	raw, err := os.ReadFile(dataFile(t, "org.mnu"))
	if err != nil {
		t.Fatal(err)
	}
	menu := new(MenuFile)
	menu.Parse(SplitLines(string(raw)))
	if len(menu.Entries) == 0 {
		t.Fatal("expected menu entries")
	}
	if menu.GetEntry("1") == nil {
		t.Fatal("missing code 1")
	}
}

func TestFixture_OptFile(t *testing.T) {
	raw, err := os.ReadFile(dataFile(t, "ws31.opt"))
	if err != nil {
		t.Fatal(err)
	}
	opt := NewOptFile()
	opt.Parse(SplitLines(string(raw)))
	if len(opt.Lines) == 0 {
		t.Fatal("expected OPT lines")
	}
}

func TestFixture_ISO2709(t *testing.T) {
	f, err := os.Open(dataFile(t, "test1.iso"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	rec, err := ReadIsoRecord(f, FromAnsi)
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Fields) == 0 {
		t.Fatal("expected fields in ISO record")
	}
}

func TestMarcRecord_RoundTrip(t *testing.T) {
	rec := NewMarcRecord()
	rec.Mfn = 5
	rec.Status = 0
	rec.Version = 2
	rec.Add(200, "").Add('a', "Title").Add('e', "Subtitle")
	rec.Add(700, "").Add('a', "Author")

	encoded := rec.Encode("\n")
	lines := SplitLines(encoded)
	// Encode ends with delimiter → trailing empty line; Decode tolerates it.
	clone := NewMarcRecord()
	clone.Decode(lines)
	if clone.Mfn != 5 || clone.Version != 2 {
		t.Fatalf("header: mfn=%d ver=%d", clone.Mfn, clone.Version)
	}
	if clone.FSM(200, 'a') != "Title" || clone.FSM(700, 'a') != "Author" {
		t.Fatalf("fields: %s", clone)
	}
}
