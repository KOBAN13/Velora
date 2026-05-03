# Velora

Velora — серверная основа на Go для сетевых игр и realtime-приложений. Сейчас проект находится на стадии транспортного и авторизационного прототипа: сервер принимает WebSocket-подключения, назначает клиентам внутренние ID, обменивается бинарными пакетами Protocol Buffers и содержит базовый слой регистрации/входа через PostgreSQL.

## Что уже есть

- WebSocket endpoint `/velora`.
- Назначение уникального `uint64` ID каждому подключенному клиенту.
- Бинарный протокол сообщений на Protocol Buffers.
- Центральный хаб подключений с каналами регистрации, отключения и broadcast.
- Состояние клиента `Connection`, которое обрабатывает запросы входа и регистрации.
- Репозиторий пользователей поверх `pgxpool`.
- Хеширование паролей через `bcrypt`.
- Настраиваемый порт запуска через CLI-флаг `-port`.

## Стек

- Go `1.26.2`
- `github.com/gorilla/websocket`
- `google.golang.org/protobuf`
- `github.com/jackc/pgx/v5`
- `github.com/joho/godotenv`
- `golang.org/x/crypto/bcrypt`

## Структура проекта

```text
.
├── main.go                              # точка входа, HTTP-сервер и endpoint /velora
├── config.env                           # локальная конфигурация окружения
├── shared
│   └── packets.proto                    # protobuf-контракт пакетов
└── server
    ├── Internal
    │   ├── idgenerator.go               # атомарный генератор ID клиентов
    │   ├── objects
    │   │   └── sharedCollection.go      # потокобезопасная коллекция клиентов
    │   └── server
    │       ├── hub.go                   # центральный хаб и подключение к БД
    │       ├── clients                  # WebSocket-клиент, чтение, запись и закрытие
    │       ├── db                       # модели и запросы пользователей
    │       └── states
    │           └── connected.go         # состояние подключения, login/register
    └── pkg
        └── packets                      # сгенерированный protobuf-код и хелперы
```

## Конфигурация

Сервер ожидает файл `config.env` и переменные окружения для PostgreSQL:

```env
DATABASE_URL=postgres://user:password@localhost:5432/velora?sslmode=disable
USERS_TABLE=users
```

Таблица пользователей должна содержать поля, которые использует репозиторий:

```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL
);
```

`USERS_TABLE` задает имя таблицы. Значение экранируется как SQL-идентификатор перед подстановкой в запросы.

## Запуск

Установите зависимости:

```bash
go mod download
```

Запустите сервер на порту `8080`:

```bash
go run .
```

Или укажите другой порт:

```bash
go run . -port 9000
```

WebSocket endpoint:

```text
ws://localhost:8080/velora
```

При запуске сервер проверяет подключение к PostgreSQL запросом `SELECT 1`. Если `config.env` отсутствует или база недоступна, процесс завершится с ошибкой.

## Протокол сообщений

Все сообщения передаются как бинарные WebSocket-фреймы с protobuf-сообщением `Packet`.

Источник схемы: [`shared/packets.proto`](./shared/packets.proto)

```proto
message Packet {
  uint64 sender_id = 1;

  oneof msg {
    ChatMessage chat = 2;
    IdMessage id = 3;
    LoginRequestMessage login_request = 4;
    RegisterRequestMessage register_request = 5;
    OkResponseMessage ok_response = 6;
    DenyResponseMessage deny_response = 7;
  }
}
```

Типы сообщений:

- `IdMessage` — служебное сообщение с ID, который сервер назначил клиенту.
- `LoginRequestMessage` — запрос входа по `username` и `password`.
- `RegisterRequestMessage` — запрос регистрации по `username` и `password`.
- `OkResponseMessage` — успешный ответ на вход или регистрацию.
- `DenyResponseMessage` — отказ с текстовой причиной.
- `ChatMessage` — тип сообщения присутствует в схеме и хелперах, но текущий state его не обрабатывает.

Если клиент отправляет пакет с `sender_id = 0`, сервер заменяет его на ID текущего WebSocket-соединения.

## Сценарий подключения

1. Клиент подключается к `/velora` по WebSocket.
2. Сервер повышает HTTP-соединение до WebSocket и регистрирует клиента в хабе.
3. Клиент получает `IdMessage` со своим серверным ID.
4. Клиент отправляет `LoginRequestMessage` или `RegisterRequestMessage`.
5. Сервер валидирует запрос, обращается к PostgreSQL и отвечает `OkResponseMessage` или `DenyResponseMessage`.
6. При отключении клиент удаляется из коллекции хаба.

## Разработка

Проверка сборки и тестов:

```bash
go test ./...
```

Перегенерация Go-кода после изменения protobuf-схемы:

```bash
protoc --go_out=server shared/packets.proto
```

Для этой команды должны быть установлены `protoc` и `protoc-gen-go`.

## Текущие ограничения

- Нет игровых комнат, сессий и репликации состояния.
- Нет прикладной маршрутизации чата, хотя `ChatMessage` уже описан в protobuf-схеме.
- Нет полноценного состояния авторизованного пользователя после успешного входа.
- Нет миграций базы данных.
- Нет проверки origin при WebSocket upgrade: `CheckOrigin` сейчас разрешает любые источники.
- Нет TLS, rate limiting и отдельного слоя валидации размера/частоты сообщений.
- Broadcast-канал есть в хабе, но текущий обработчик состояния не использует его для пользовательских сообщений.

## Лицензия

См. [`LICENSE`](./LICENSE).
