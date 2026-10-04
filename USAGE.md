# GoIrbis — возможности и примеры

Клиентская библиотека для работы с сервером **ИРБИС64** на языке Go.
Порт ManagedIrbis: чистый TCP-протокол, без `irbis64_client.dll` и без внешних зависимостей.

Пакет: `src/irbis` (`package irbis`).

> Сейчас репозиторий в стиле GOPATH (без `go.mod`). Локально удобнее подключать пакет относительным импортом; после появления module path — через `go get`.

```go
import "github.com/ivm97/GoIrbis/src/irbis" // после go.mod
// или
import "./src/irbis"                       // локально / GOPATH
```

Требования: сервер ИРБИС64 (от ~2014), Go 1.12+ (проверено также на современных версиях).

---

## Основные возможности

| Область | Что умеет |
|--------|-----------|
| Подключение | TCP-клиент, строка подключения, INI с сервера, `NoOp` |
| Записи | чтение / запись / удаление / восстановление, пакетные операции |
| Поиск | простой, расширенный, с лимитом 32k+, count, search+read |
| Форматы | `@brief`, произвольный PFT, Unicode в формате |
| Словарь | термины, постинги, префиксы |
| Файлы сервера | текст, меню, INI, OPT, PAR, TRE, список файлов |
| Администрирование | БД, пользователи, процессы, статистика, GBL |
| Локальный доступ | прямое чтение MST/XRF (`DirectAccess`) |
| Утилиты | ISO2709, builder поисковых выражений, CP1251 |

Клиент **однопоточный**: один `Connection` — один поток. Для параллелизма нужны несколько подключений (если позволяет лицензия сервера). Keep-alive сервер сам не шлёт — вызывайте `NoOp` по таймеру.

---

## Подключение

Значения по умолчанию: `127.0.0.1:6666`, база `IBIS`, АРМ `C` (каталогизатор).

```go
package main

import (
	"fmt"
	"log"

	"github.com/ivm97/GoIrbis/src/irbis"
)

func main() {
	conn := irbis.NewConnection()
	conn.Host = "localhost"
	conn.Port = 6666
	conn.Username = "librarian"
	conn.Password = "secret"
	conn.Database = "IBIS"
	conn.Workstation = irbis.CATALOGER // "C"

	if !conn.Connect() {
		log.Fatal("не удалось подключиться:", irbis.DescribeError(conn.LastError))
	}
	defer conn.Disconnect()

	fmt.Println("версия сервера:", conn.ServerVersion)
	fmt.Println("интервал (мин):", conn.Interval)

	// настройки клиента с сервера
	fmtMenu := conn.Ini.GetValue("Main", "FmtMnu", "FMT31.MNU")
	fmt.Println("FmtMnu:", fmtMenu)

	conn.NoOp() // подтверждение сессии
}
```

Строка подключения:

```go
conn := irbis.NewConnection()
conn.ParseConnectionString(
	"host=192.168.1.4;port=5555;username=itsme;password=secret;db=IBIS;arm=C;",
)
if !conn.Connect() {
	log.Fatal(irbis.DescribeError(conn.LastError))
}
defer conn.Disconnect()

fmt.Println(conn.ToConnectionString())
```

Ключи: `host|server|address`, `port`, `user|username|name|login`, `pwd|password`, `db|database|catalog`, `arm|workstation`.

Коды АРМ: `A` админ, `C` каталогизатор, `M` комплектатор, `R` читатель, `B` книговыдача, `K` книгообеспеченность.

---

## Поиск

Поисковое выражение — синтаксис ИРБИС; кавычки обычно нужны:

```go
found := conn.Search(`"A=ПУШКИН$"`)
fmt.Println("найдено (≤32k):", len(found))

count := conn.SearchCount(`"A=ПУШКИН$"`)
fmt.Println("всего по запросу:", count)

all := conn.SearchAll(`"A=ПУШКИН$"`) // несколько запросов к серверу
fmt.Println("все MFN:", len(all))
```

Поиск с чтением записей:

```go
records := conn.SearchRead(`"A=ПУШКИН$"`, 50)
for _, rec := range records {
	fmt.Println(rec.Mfn, rec.FSM(200, 'a'))
}

one := conn.SearchSingleRecord(`"I=65.304.13-772296"`)
if one == nil {
	fmt.Println("не найдено")
}
```

Расширенный поиск (лимит, формат, диапазон MFN):

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

Builder выражений:

```go
expr := irbis.Author("Пушкин$").And(irbis.Year("2020")).String()
// примерно: ("A=Пушкин$" * "G=2020")
found := conn.Search(expr)
```

Хелперы: `Author`, `Title`, `Keyword`, `Year`, `Language`, `Publisher`, `Subject`, `Udc`, `Bbk`, `Number`, `Mhr`, …; операторы `And` / `Or` / `Not` / `SameField` / `SameRepeat`.

Частые префиксы: `A=` автор, `T=` заглавие, `K=` ключевые слова, `I=` шифр, `IN=` инвентарный номер, `G=` год, `J=` язык.

---

## Записи (MarcRecord)

Чтение:

```go
rec := conn.ReadRecord(123)
if rec == nil {
	log.Fatal("запись не прочитана:", irbis.DescribeError(conn.LastError))
}
fmt.Println("заглавие:", rec.FSM(200, 'a'))
fmt.Println("автор:", rec.FSM(700, 'a'))

batch := conn.ReadRecords([]int{1, 2, 3})
old := conn.ReadRecordVersion(123, 3)
```

Создание и запись:

```go
conn.Database = "SANDBOX"

rec := irbis.NewMarcRecord()
rec.Add(700, "").
	Add('a', "Миронов").
	Add('b', "А. В.").
	Add('g', "Алексей Владимирович")
rec.Add(200, "").
	Add('a', "Работа с ИРБИС64").
	Add('e', "руководство пользователя")
rec.Add(210, "").
	Add('a', "Иркутск").
	Add('c', "ИРНИТУ").
	Add('d', "2019")
rec.Add(920, "PAZK")

maxMfn := conn.WriteRecord(rec) // после записи сервер может прогнать AUTOIN.GBL
fmt.Println("MFN:", rec.Mfn, "max MFN:", maxMfn)
```

Пакетная запись, удаление, восстановление:

```go
ok := conn.WriteRecords(records)

conn.DeleteRecord(123)                 // логическое удаление
restored := conn.UndeleteRecord(123)
```

Полезные методы записи:

- `FM(tag)` / `FSM(tag, code)` — первое поле / подполе  
- `FMA` / `FSMA` — все значения  
- `GetField` / `GetFields` / `HaveField`  
- `SetField` / `SetSubfield` / `RemoveField`  
- `IsDeleted`, `Clone`, `Encode` / `Decode`

Статусы (битовые флаги): `LOGICALLY_DELETED`, `PHYSICALLY_DELETED`, `LOCKED_RECORD`, `NON_ACTUALIZED`, …

---

## Форматирование

```go
text := conn.FormatMfn(irbis.BRIEF_FORMAT, 123)
fmt.Println(text)

// произвольный PFT (Unicode допустим)
text = conn.FormatMfn("'Автор: ', v700^a", 123)

lines := conn.FormatRecords(irbis.BRIEF_FORMAT, []int{1, 2, 3})

// запись ещё не на сервере — с клиента
draft := irbis.NewMarcRecord()
draft.Add(200, "").Add('a', "Черновик")
fmt.Println(conn.FormatRecord(irbis.BRIEF_FORMAT, draft))
```

Константы форматов: `BRIEF_FORMAT` (`@brief`), `ALL_FORMAT`, `IBIS_FORMAT`, `INFO_FORMAT`, `OPTIMIZED_FORMAT`.

---

## Словарь: термины и постинги

```go
terms := conn.ReadTerms("A=ПУШ", 20)
for _, t := range terms {
	fmt.Println(t.Text, t.Count)
}

langs := conn.ListTerms("J=")
fmt.Println(langs)

pp := irbis.NewPostingParameters()
pp.Term = "J=CHI"
pp.NumberOfPostings = 100
fmt.Println(conn.ReadPostings(pp))

fmt.Println(conn.GetRecordPostings(2, "A=$"))
```

---

## Файлы на сервере

Спецификация вида `путь.база.имя` (например `3.IBIS.WS31.OPT`).

```go
content := conn.ReadTextFile("3.IBIS.WS.OPT")
lines := conn.ReadTextLines("3.IBIS.brief.pft")

menu := conn.ReadMenuFile("3.IBIS.FORMATW.MNU")
ini := conn.ReadIniFile("...")
opt := conn.ReadOptFile("3.IBIS.WS31.OPT")
par := conn.ReadParFile("1..IBIS.PAR")
tree := conn.ReadTreeFile("3.IBIS.II.TRE")

files := conn.ListFiles("3.IBIS.brief.*", "3.IBIS.a*.pft")

conn.WriteTextFile("3.IBIS.my.txt", "строка1\nстрока2\n")
conn.DeleteFile("3.IBIS.my.txt")
```

---

## Сведения о сервере и администрирование

```go
ver := conn.GetServerVersion()
fmt.Println(ver.Organization, ver.Version)

stat := conn.GetServerStat()
fmt.Println(stat)

fmt.Println(conn.ListProcesses())
fmt.Println(conn.ListDatabases("")) // или спецификация dbnam*.mnu
fmt.Println(conn.GetDatabaseInfo("IBIS"))
fmt.Println(conn.GetMaxMfn("IBIS"))
fmt.Println(conn.GetUserList())

// осторожно: меняет состояние сервера
conn.CreateDatabase("TEST", "Тестовая", true)
conn.CreateDictionary("TEST")
conn.ActualizeDatabase("TEST")
conn.TruncateDatabase("TEST")
conn.DeleteDatabase("TEST")
conn.UnlockDatabase("IBIS")
```

Глобальная корректировка — через `GlobalCorrection(*GblSettings)`.

---

## Прямой доступ к файлам БД (без сервера)

Если есть файлы MST/XRF на диске:

```go
access, err := irbis.OpenDatabase("/path/to/ibis.mst")
if err != nil {
	log.Fatal(err)
}
defer access.Close()

fmt.Println("max MFN:", access.GetMaxMfn())
rec, err := access.ReadRecord(1)
if err != nil {
	log.Fatal(err)
}
fmt.Println(rec.FSM(200, 'a'))
```

Также есть разбор ISO2709 (`ReadIsoRecord`), экспорт в plain text (`ToPlainText` / `ExportPlainText`).

---

## Обработка ошибок

Большинство методов возвращают `bool`, `nil` или пустой результат; код — в `conn.LastError`.

```go
if !conn.Connect() {
	fmt.Println(irbis.DescribeError(conn.LastError))
	return
}

rec := conn.ReadRecord(999999)
if rec == nil {
	fmt.Println(irbis.DescribeError(conn.LastError))
}

// аварийный выход (log.Fatal) — только для простых скриптов
conn.FailOnError()
```

Типичные коды: `-3333` клиент не в списке, `-4444` неверный пароль, `-5555` файл не найден, `-602` запись заблокирована, `-608` конфликт версии.

---

## Примеры в репозитории

| Файл | Назначение |
|------|------------|
| `examples/Example1.go` | подключение, поиск, чтение, `@brief` |
| `examples/Example2.go` | создание и запись 10 записей |
| `examples/icqbot.go` | бот-пример поверх библиотеки |
| `src/SafeExperiments.go` | «дымовой» обход многих API-методов |

Запуск примера (GOPATH-режим, из корня репозитория):

```bash
GO111MODULE=off GOPATH="$PWD" go run examples/Example1.go
```

Тесты пакета:

```bash
GO111MODULE=off GOPATH="$PWD" go test -v ./src/irbis
```

---

## Краткая шпаргалка API `Connection`

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
