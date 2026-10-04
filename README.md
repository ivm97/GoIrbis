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
	"context"
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

	found, err := client.Search(ctx, `"A=Byron, George$"`)
	if err != nil {
		log.Fatal(err)
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

По умолчанию: `127.0.0.1:6666`, база `IBIS`, АРМ каталогизатора (`C`).  
`*Client` на каждый вызов делает login → команда → logout; параллельные запросы безопасны (отдельные сессии, без мьютексов). Для длинной сессии есть `Connection`.

Таймауты задаются в `irbis.Config`. По умолчанию на команду нет жёсткого IO-лимита (ИРБИС часто отвечает долго): ограничивайте запрос через `context`. Срыв дедлайна обрывает только ожидание на клиенте, сервер команду сам не отменяет.

Ошибки — тип `irbis.Error` с кодами (`CodeWrongPassword`, `CodeNetwork`, …) и текстами из протокола ИРБИС; удобно проверять через `errors.Is(err, irbis.ErrWrongPassword)`.

Длинную сессию по-прежнему даёт `Connection`. Пока она жива, периодически вызывайте `NoOp`, иначе сервер может её сбросить.

```go
client := irbis.NewClient(irbis.Config{
	Host:     "localhost",
	Username: "librarian",
	Password: "secret",
	// IOTimeout: 3 * time.Minute, // если нужен жёсткий потолок на одну TCP-команду
})
```

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
