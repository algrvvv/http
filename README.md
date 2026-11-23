# http — HTTP-клиент для терминала

## Установка

Установите Go и выполните команду из любого каталога.

```sh
go install github.com/algrvvv/http/cmd/http@latest
```

Go установит бинарник `http` в каталог `GOBIN`. Если он не задан, используется `$(go env GOPATH)/bin`. Если оболочка не находит команду, добавьте каталог установки в `PATH`.

```sh
http_bin_dir="$(go env GOBIN)"
export PATH="$PATH:${http_bin_dir:-$(go env GOPATH)/bin}"
http --version
```

Чтобы сохранить настройку для новых терминалов, добавьте обе строки настройки `PATH` в `~/.zshrc` или `~/.bashrc`.

Для установки из текущей ветки выполните `go install ./cmd/http` из корня репозитория. Для локальной сборки используйте `make build`, после чего утилиту можно запустить как `./bin/http`.

## Запросы

```sh
http GET localhost:8000/api/data -H 'Accept: application/json' -t 5
http POST localhost:8000/api/data -J message=hello 'author=Иван Иванов'
http POST localhost:8000/api/data -J -b '{"items":[{"count":2,"active":true}]}'
http POST localhost:8000/api/data -J -b @payload.json
cat payload.json | http POST localhost:8000/api/data -J -b @-
```

С `-J` CLI выставляет `Content-Type: application/json`, если он не задан через `-H`. Позиционные пары образуют JSON-объект со строковыми значениями. Для чисел, булевых значений, массивов и вложенных объектов передайте готовый JSON через `-b`. Используйте один источник тела запроса за вызов.

`-H` можно повторять. Для настроек запроса также есть `--cookies 'session=value'`, `--user-agent 'example-client'` и `--proxy http://localhost:8080`.

## Ответ и диагностика

```sh
http GET localhost:8000/api/data --pretty -A
http GET localhost:8000/api/data --body-only --fail | jq .
http GET localhost:8000/api/data -r --verbose
http GET localhost:8000/api/data -W
```

- По умолчанию выводятся статус, итоговый URL, время до заголовков, основные заголовки и тело. `-A` показывает все заголовки, `-W` пропускает чтение тела, а `--pretty` форматирует JSON.
- `--body-only` пишет исходные байты ответа в stdout. Ошибки и вывод `--verbose` идут в stderr.
- Ошибки запроса возвращают ненулевой код выхода. `--fail` дополнительно возвращает ненулевой код при HTTP 4xx и 5xx, сохраняя вывод ответа.
- `-r` включает переходы по редиректам. `--verbose` печатает пройденные переходы и общее время, включая чтение тела, если оно включено.
- `-I` отключает проверку TLS-сертификата для локальной отладки. Все флаги доступны через `http --help`.

## Алиасы хостов

Сохраните базовый URL один раз и используйте имя в следующих вызовах.

```sh
http set myapp localhost:8000/api/v1/
http POST myapp/login -J username=demo
http GET 'myapp/users?page=2'
```

Первый запрос после сохранения уйдёт на `http://localhost:8000/api/v1/login`. Алиасы хранятся в пользовательском каталоге конфигурации в подкаталоге `http/hosts`. Повторный `http set` обновляет адрес, имена регистронезависимы. Явный URL со схемой используется напрямую.
