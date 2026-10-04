# GoIrbis

Клиентская библиотека на Go для работы с сервером **ИРБИС64**.

Изначально это порт ManagedIrbis, который написал **Алексей Миронов** ([amironov73](https://github.com/amironov73)). Оригинальный репозиторий он удалил; здесь живёт форк с доработками под современный Go. Авторство исходного кода — за ним (см. также `LICENSE`, Copyright 2019 Alexey Mironov).

Библиотека говорит с сервером по его родному TCP-протоколу. Не нужна `irbis64_client.dll`, нет CGO и нет внешних зависимостей у пакета `irbis`. На Windows и Linux ведёт себя одинаково. По объёму сейчас около 5.5k строк в самом пакете — поиск, чтение и запись MARC-записей, словарь, файлы на сервере, админские операции и прямое чтение MST/XRF с диска.

Ожидается Go 1.22+ и сервер ИРБИС64 примерно от 2014 года.

## Установка

```bash
go get github.com/ivm97/GoIrbis/irbis@latest
```

```go
import "github.com/ivm97/GoIrbis/irbis"
```

## Быстрый старт

```go
package main

import (
	"fmt"
	"log"

	"github.com/ivm97/GoIrbis/irbis"
)

func main() {
	conn := irbis.NewConnection()
	conn.Host = "localhost"
	conn.Username = "librarian"
	conn.Password = "secret"

	if !conn.Connect() {
		log.Fatal(irbis.DescribeError(conn.LastError))
	}
	defer conn.Disconnect()

	found := conn.Search(`"A=Byron, George$"`)
	for _, mfn := range found {
		rec := conn.ReadRecord(mfn)
		if rec == nil {
			continue
		}
		fmt.Println(rec.FSM(200, 'a'))
		fmt.Println(conn.FormatMfn(irbis.BRIEF_FORMAT, mfn))
	}
}

```

По умолчанию клиент стучится на `127.0.0.1:6666`, база `IBIS`, АРМ каталогизатора (`C`). Строку подключения можно разобрать через `ParseConnectionString`.

Один экземпляр `Connection` рассчитан на последовательную работу. Если нужны параллельные запросы — открывайте отдельное подключение на каждый поток или HTTP-запрос (с логином заново), насколько позволяет лицензия сервера. Пока сессия жива, периодически вызывайте `NoOp`, иначе сервер может её сбросить.

## Что умеет

Поиск по выражениям ИРБИС (в том числе больше 32 тысяч записей через `SearchAll`), чтение и сохранение записей, расформатирование на сервере (`@brief` и произвольный PFT), работа со словарём и постингами, чтение меню/INI/OPT/PAR/TRE, создание и удаление баз, список пользователей и процессов. Если сервер недоступен, а файлы базы есть на диске — `DirectAccess` читает MST/XRF напрямую.

Подробности и рецепты — в [USAGE.md](USAGE.md). Живые примеры лежат в `examples/`: поиск и чтение, запись в `SANDBOX`, офлайн-разбор ISO и MST, плюс старый icq-бот как иллюстрация встраивания.

```bash
go test ./irbis/...
go run ./examples/search_and_read
```

## Структура репозитория

`irbis/` — сам модуль, его и подключают в проекты.  
`examples/` — демо, не часть публичного API.  
`data/` — небольшие фикстуры для офлайн-примеров.

## Лицензия

MIT. Исходный copyright — Alexey Mironov, 2019.
