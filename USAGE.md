# GoIrbis — usage

Go client for the **IRBIS64** library server.
Pure TCP protocol — no `irbis64_client.dll`, no CGO, no third-party deps in the library module.

Package import:

```go
import "github.com/ivm97/GoIrbis/irbis"
```

```bash
go get github.com/ivm97/GoIrbis/irbis@latest
```

Requirements: Go 1.22+, IRBIS64 server (≈2014+).

---

## Capabilities

| Area | Features |
|------|----------|
| Connection | TCP client, connection string, server INI, `NoOp` |
| Records | read / write / delete / undelete, batch ops |
| Search | simple, extended, >32k via paging, count, search+read |
| Format | `@brief`, custom PFT, Unicode in format |
| Dictionary | terms, postings, prefixes |
| Server files | text, menu, INI, OPT, PAR, TRE, file list |
| Admin | databases, users, processes, stats, GBL |
| Direct access | local MST/XRF (`DirectAccess`) |
| Utilities | ISO2709, search expression builder, CP1251 |

Prefer `*Client` for HTTP/API handlers: one logical session per call, `context` for cancellation.
Use `Connection` when you need a long-lived session and many commands without re-login.
Keep-alive is not automatic — call `NoOp` on a timer while a long session stays open.

### Timeouts

Client-side only. If a deadline fires, Go closes the local TCP wait; IRBIS is **not** told to abort and may still finish the command on the server.

- `DialTimeout` — TCP connect. Default `10s` when left zero in `Config`.
- `IOTimeout` — one command write+read when context has no deadline. Default `0` = wait without transport limit (recommended for slow IRBIS). Set explicitly for hard caps.
- `context` deadline/cancel — always honored and is the usual way to bound a request in services.

```go
client := irbis.NewClient(irbis.Config{
	Host:        "localhost",
	Username:    "librarian",
	Password:    "secret",
	Database:    "IBIS",
	DialTimeout: 10 * time.Second,
	IOTimeout:   3 * time.Minute, // optional hard cap per TCP command
})
```

---

## Client (recommended for services)

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ivm97/GoIrbis/irbis"
)

func main() {
	client := irbis.NewClient(irbis.Config{
		Host:     "localhost",
		Username: "librarian",
		Password: "secret",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	found, err := client.Search(ctx, `"A=ПУШКИН$"`)
	if err != nil {
		if errors.Is(err, irbis.ErrWrongPassword) {
			log.Fatal("bad credentials")
		}
		log.Fatal(err) // also: irbis.CodeOf(err), irbis.DescribeError(...)
	}

	for _, mfn := range found {
		rec, err := client.ReadRecord(ctx, mfn)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(rec.FSM(200, 'a'))
	}
}
```

`Do` runs arbitrary work inside one short session:

```go
err := client.Do(ctx, func(conn *irbis.Connection) error {
	_ = conn.NoOp()
	return nil
})
```

Methods: `Do`, `Search`, `SearchCount`, `SearchAll`, `SearchRead`, `ReadRecord`, `ReadRecords`, `WriteRecord`, `FormatMfn`, `GetMaxMfn`, `ReadTerms`, `ReadTextFile`, `NoOp`.

---

## Connection (long session)

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ivm97/GoIrbis/irbis"
)

func main() {
	conn := irbis.NewConnection()
	conn.Host = "localhost"
	conn.Username = "librarian"
	conn.Password = "secret"

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := conn.ConnectContext(ctx); err != nil {
		log.Fatal(err)
	}
	defer conn.DisconnectContext(context.Background())

	fmt.Println("server version:", conn.ServerVersion)
	fmt.Println("interval (min):", conn.Interval)

	fmtMenu := conn.Ini.GetValue("Main", "FmtMnu", "FMT31.MNU")
	fmt.Println("FmtMnu:", fmtMenu)

	conn.NoOp()
}
```

Legacy `Connect()` / `Disconnect()` / `Execute()` still work; they use `context.Background()`.

Connection string:

```go
conn := irbis.NewConnection()
conn.ParseConnectionString(
	"host=192.168.1.4;port=5555;username=itsme;password=secret;db=IBIS;arm=C;",
)
if !conn.Connect() {
	log.Fatal(irbis.DescribeError(conn.LastError))
}
defer conn.Disconnect()
```

Keys: `host|server|address`, `port`, `user|username|name|login`, `pwd|password`, `db|database|catalog`, `arm|workstation`.

Workstation codes: `A` admin, `C` cataloger, `M` acquisitions, `R` reader, `B` circulation, `K` provision.

---

## Search

```go
found := conn.Search(`"A=ПУШКИН$"`)
fmt.Println("found (≤32k):", len(found))

count := conn.SearchCount(`"A=ПУШКИН$"`)
all := conn.SearchAll(`"A=ПУШКИН$"`) // multiple server round-trips
```

Search and load records:

```go
records := conn.SearchRead(`"A=ПУШКИН$"`, 50)
for _, rec := range records {
	fmt.Println(rec.Mfn, rec.FSM(200, 'a'))
}

one := conn.SearchSingleRecord(`"I=65.304.13-772296"`)
if one == nil {
	fmt.Println("not found")
}
```

Extended search:

```go
params := irbis.NewSearchParameters()
params.Expression = `"A=ПУШКИН$"`
params.Format = irbis.BRIEF_FORMAT
params.NumberOfRecords = 5
params.FirstRecord = 1

for _, line := range conn.SearchEx(params) {
	fmt.Println(line.Mfn, line.Description)
}
```

Expression builder:

```go
expr := irbis.Author("Пушкин$").And(irbis.Year("2020")).String()
found := conn.Search(expr)
```

Helpers: `Author`, `Title`, `Keyword`, `Year`, `Language`, `Publisher`, `Subject`, `Udc`, `Bbk`, `Number`, `Mhr`, …  
Operators: `And` / `Or` / `Not` / `SameField` / `SameRepeat`.

Common prefixes: `A=` author, `T=` title, `K=` keywords, `I=` document index, `IN=` inventory, `G=` year, `J=` language.

---

## Records (MarcRecord)

```go
rec := conn.ReadRecord(123)
if rec == nil {
	log.Fatal(irbis.DescribeError(conn.LastError))
}
fmt.Println(rec.FSM(200, 'a'))

batch := conn.ReadRecords([]int{1, 2, 3})
old := conn.ReadRecordVersion(123, 3)
```

Create and write:

```go
conn.Database = "SANDBOX"

rec := irbis.NewMarcRecord()
rec.Add(700, "").
	Add('a', "Миронов").
	Add('b', "А. В.").
	Add('g', "Алексей Владимирович")
rec.Add(200, "").
	Add('a', "Working with IRBIS64").
	Add('e', "user guide")
rec.Add(210, "").
	Add('a', "Irkutsk").
	Add('c', "IRNITU").
	Add('d', "2019")
rec.Add(920, "PAZK")

maxMfn := conn.WriteRecord(rec) // server may run AUTOIN.GBL
fmt.Println("MFN:", rec.Mfn, "max MFN:", maxMfn)
```

```go
ok := conn.WriteRecords(records)
conn.DeleteRecord(123)
restored := conn.UndeleteRecord(123)
```

Useful methods: `FM` / `FSM` / `FMA` / `FSMA`, `GetField` / `GetFields`, `SetField` / `SetSubfield`, `RemoveField`, `IsDeleted`, `Clone`, `Encode` / `Decode`.

Status flags: `LOGICALLY_DELETED`, `PHYSICALLY_DELETED`, `LOCKED_RECORD`, `NON_ACTUALIZED`, …

---

## Formatting

```go
text := conn.FormatMfn(irbis.BRIEF_FORMAT, 123)
text = conn.FormatMfn("'Author: ', v700^a", 123)
lines := conn.FormatRecords(irbis.BRIEF_FORMAT, []int{1, 2, 3})

draft := irbis.NewMarcRecord()
draft.Add(200, "").Add('a', "Draft")
fmt.Println(conn.FormatRecord(irbis.BRIEF_FORMAT, draft))
```

Format constants: `BRIEF_FORMAT`, `ALL_FORMAT`, `IBIS_FORMAT`, `INFO_FORMAT`, `OPTIMIZED_FORMAT`.

---

## Dictionary

```go
terms := conn.ReadTerms("A=ПУШ", 20)
langs := conn.ListTerms("J=")

pp := irbis.NewPostingParameters()
pp.Term = "J=CHI"
pp.NumberOfPostings = 100
fmt.Println(conn.ReadPostings(pp))
fmt.Println(conn.GetRecordPostings(2, "A=$"))
```

---

## Server files

Specification form: `path.database.name` (e.g. `3.IBIS.WS31.OPT`).

```go
content := conn.ReadTextFile("3.IBIS.WS.OPT")
menu := conn.ReadMenuFile("3.IBIS.FORMATW.MNU")
opt := conn.ReadOptFile("3.IBIS.WS31.OPT")
par := conn.ReadParFile("1..IBIS.PAR")
tree := conn.ReadTreeFile("3.IBIS.II.TRE")
files := conn.ListFiles("3.IBIS.brief.*", "3.IBIS.a*.pft")

conn.WriteTextFile("3.IBIS.my.txt", "line1\nline2\n")
conn.DeleteFile("3.IBIS.my.txt")
```

---

## Server info and admin

```go
ver := conn.GetServerVersion()
stat := conn.GetServerStat()
fmt.Println(conn.ListProcesses())
fmt.Println(conn.ListDatabases(""))
fmt.Println(conn.GetDatabaseInfo("IBIS"))
fmt.Println(conn.GetMaxMfn("IBIS"))
fmt.Println(conn.GetUserList())
```

Destructive / privileged ops: `CreateDatabase`, `DeleteDatabase`, `TruncateDatabase`, `Actualize*`, `UnlockDatabase`, `GlobalCorrection`.

---

## Direct DB access (no server)

```go
access, err := irbis.OpenDatabase("/path/to/ibis.mst")
if err != nil {
	log.Fatal(err)
}
defer access.Close()

rec, err := access.ReadRecord(1)
if err != nil {
	log.Fatal(err)
}
fmt.Println(rec.FSM(200, 'a'))
```

Also: `ReadIsoRecord`, `ToPlainText` / `ExportPlainText`.

---

## Errors

Context-aware APIs (`Client`, `ConnectContext`, …) return `error` as `*irbis.Error`:
numeric IRBIS code + English description.

```go
err := client.Search(ctx, expr)
if errors.Is(err, irbis.ErrUnregisteredClient) { /* ... */ }
if errors.Is(err, irbis.ErrNetwork) { /* dial/timeout/cancel */ }

code := irbis.CodeOf(err)                 // e.g. irbis.CodeWrongPassword
text := irbis.DescribeError(code)         // same text as err.Error() without cause
```

Constants: `CodeWrongPassword`, `CodeUnregisteredClient`, `CodeRecordLocked`, `CodeNetwork`, …
Sentinels for `errors.Is`: `ErrWrongPassword`, `ErrNetwork`, …

Legacy `Connection` methods still use `bool`/`nil` + `conn.LastError`; prefer `conn.Err()` after a failed call.

`FailOnError()` calls `log.Fatal` — only for tiny scripts.

---

## Examples in this repo

See [examples/README.md](examples/README.md).

```bash
go test ./irbis/...
go run ./examples/search_and_read
go run ./examples/direct_access
```

---

## Connection API cheat sheet

```
Connect / Disconnect / NoOp / ParseConnectionString / ToConnectionString
Search / SearchAll / SearchCount / SearchEx / SearchRead / SearchSingleRecord
ReadRecord / ReadRecords / ReadRecordVersion / ReadRawRecord
WriteRecord / WriteRecords / WriteRawRecord
DeleteRecord / UndeleteRecord
FormatMfn / FormatRecord / FormatRecords
ReadTerms / ReadTermsEx / ListTerms / ReadPostings / GetRecordPostings
ReadTextFile / ReadTextLines / WriteTextFile / ListFiles / DeleteFile
ReadMenuFile / ReadIniFile / ReadOptFile / ReadParFile / ReadTreeFile
GetMaxMfn / GetDatabaseInfo / ListDatabases
GetServerVersion / GetServerStat / ListProcesses / GetUserList
CreateDatabase / DeleteDatabase / TruncateDatabase / Actualize*
UnlockDatabase / UnlockRecords / GlobalCorrection / PrintTable
```
