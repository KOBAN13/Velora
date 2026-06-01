# Velora

Velora - серверная основа на Go для сетевых игр и realtime-приложений. Проект принимает WebSocket-подключения, обменивается бинарными пакетами Protocol Buffers, авторизует пользователей через PostgreSQL и содержит in-memory lobby manager для игровых комнат.

## Что уже есть

- WebSocket endpoint `/velora`.
- Назначение уникального `uint64` ID каждому подключенному клиенту.
- Бинарный протокол сообщений на Protocol Buffers.
- Центральный хаб подключений с регистрацией, отключением и broadcast.
- Состояние `Connection` для регистрации, входа и базового chat/broadcast.
- Состояние `Authenticated` для lobby-команд после успешного входа.
- PostgreSQL-репозиторий пользователей поверх `pgxpool`.
- Хеширование паролей через `bcrypt`.
- In-memory lobby manager с комнатами, владельцем комнаты, ready-флагами и стартом матча.
- Загрузка стартовых игровых параметров из Google Sheets.
- Protobuf snapshots для состояния комнаты и краткого списка комнат.
- Настраиваемый порт запуска через CLI-флаг `-port`.

## Стек

- Go `1.26.2`
- `github.com/gorilla/websocket`
- `google.golang.org/protobuf`
- `github.com/jackc/pgx/v5`
- `github.com/joho/godotenv`
- `golang.org/x/crypto/bcrypt`
- `google.golang.org/api/sheets/v4`

## Структура проекта

```text
.
├── main.go                              # точка входа, HTTP-сервер и endpoint /velora
├── Makefile                             # proto generation и deploy-команды
├── docker-compose.yml                   # production-контейнер
├── shared
│   └── packets.proto                    # protobuf-контракт пакетов
└── server
    ├── Internal
    │   ├── idgenerator.go               # генератор ID
    │   ├── objects
    │   │   └── sharedCollection.go      # потокобезопасная коллекция
    │   └── server
    │       ├── hub.go                   # центральный хаб, клиенты, lobby и БД
    │       ├── clients                  # WebSocket-клиент, чтение, запись и закрытие
    │       ├── config                   # загрузка env и игровых параметров из Google Sheets
    │       ├── contracts                # интерфейсы hub/client/lobby/state
    │       ├── db                       # модели и запросы пользователей
    │       ├── lobby                    # комнаты, игроки, snapshots и match start
    │       └── states                   # Connection и Authenticated states
    └── pkg
        └── packets                      # сгенерированный protobuf-код и хелперы
```

## Конфигурация

Сервер загружает `config.env` при старте. В текущей реализации отсутствие файла считается ошибкой запуска.

```env
DATABASE_URL=postgres://user:password@localhost:5432/velora?sslmode=disable
USERS_TABLE=users
GOOGLE_SHEETS_SPREADSHEET_ID=spreadsheet-id
GOOGLE_SHEETS_CONFIG_SHEET_PLAYER_PARAMETERS=PlayerParameters_Default
GOOGLE_SHEETS_CONFIG_SHEET_CORE_ENTITY=Core_Entity
GOOGLE_SHEETS_CONFIG_SHEET_NUTRIENT=Nutrient
GOOGLE_SHEETS_CONFIG_SHEET_PLAYER_WALLS=Walls
GOOGLE_SERVICE_ACCOUNT_FILE=./service-account.json
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

Игровой конфиг собирается при старте через `config.NewAppConfig`. Сервер создает Google Sheets client по `GOOGLE_SERVICE_ACCOUNT_FILE`, затем читает четыре листа из таблицы `GOOGLE_SHEETS_SPREADSHEET_ID`:

- `GOOGLE_SHEETS_CONFIG_SHEET_PLAYER_PARAMETERS` - параметры стартовой player cell.
- `GOOGLE_SHEETS_CONFIG_SHEET_CORE_ENTITY` - параметры core entity.
- `GOOGLE_SHEETS_CONFIG_SHEET_NUTRIENT` - параметры nutrient.
- `GOOGLE_SHEETS_CONFIG_SHEET_PLAYER_WALLS` - параметры walls.

Файл service account должен быть доступен по пути из `GOOGLE_SERVICE_ACCOUNT_FILE`, а у service account должен быть доступ на чтение указанной Google Sheets таблицы.

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

При запуске сервер загружает игровой конфиг из Google Sheets, создает пул PostgreSQL и проверяет подключение через `Ping` и `SELECT 1`. Если `config.env`, обязательные env-переменные, service account, Google Sheets таблица или база недоступны, процесс завершится с ошибкой.

## Деплой

Для деплоя с локальной машины используйте:

```bash
make deploy-remote
```

Команда синхронизирует проект в `root@apisfs.ru:/opt/velora`, удаляет на сервере файлы, которых больше нет локально, и затем пересобирает контейнер. Серверный `config.env` при этом не трогается.

Если изменения уже загружены на сервер, пересоберите и пересоздайте контейнер прямо на сервере:

```bash
make deploy
```

Команда запускает:

```bash
docker compose up -d --build --force-recreate --remove-orphans
```

`config.env` не копируется в Docker-образ. В production он передается контейнеру через `env_file` в `docker-compose.yml`, поэтому файл должен лежать рядом с `docker-compose.yml` на сервере. Контейнер слушает порт `8080`, наружу он проброшен как `9999`.

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
    CreateRoomRequestMessage create_room_request = 8;
    JoinRoomRequestMessage join_room_request = 9;
    LeaveRoomRequestMessage leave_room_request = 10;
    ReadyRequestMessage ready_request = 11;
    RoomStateSnapshotMessage room_state_snapshot = 12;
    MatchStartMessage match_started = 13;
    StartGameRequestMessage start_game = 14;
    RoomListRequestMessage room_list = 15;
    RoomSummaryMessage room_summary_message = 16;
    RoomListSnapshotMessage room_list_snapshot = 17;
  }
}
```

Connection/auth сообщения:

- `IdMessage` - служебное сообщение с ID, который сервер назначил клиенту.
- `LoginRequestMessage` - запрос входа по `username` и `password`.
- `RegisterRequestMessage` - запрос регистрации по `username` и `password`.
- `ChatMessage` - сообщение для базовой маршрутизации через `Connection`.

Lobby сообщения:

- `CreateRoomRequestMessage` - создать комнату, поля `roomName` и `maxPlayer`.
- `JoinRoomRequestMessage` - войти в комнату по `roomId`.
- `LeaveRoomRequestMessage` - выйти из текущей комнаты.
- `ReadyRequestMessage` - изменить готовность текущего игрока.
- `StartGameRequestMessage` - стартовать матч владельцем комнаты.
- `RoomListRequestMessage` - запросить список комнат.

Snapshots и ответы:

- `RoomStateSnapshotMessage` - состояние одной комнаты: `roomId`, `maxPlayer`, `status`, список игроков.
- `RoomPlayerMessage` - игрок комнаты: `userId`, `clientId`, `username`, `isReady`, `owner`.
- `RoomListSnapshotMessage` - краткий список комнат.
- `RoomSummaryMessage` - summary комнаты: `name`, `roomId`, `playersCount`, `maxPlayer`, `status`.
- `MatchStartMessage` - уведомление о старте матча с `roomId` и `matchId`.
- `OkResponseMessage` - успешный ответ на команду.
- `DenyResponseMessage` - отказ с текстовой причиной.

Если клиент отправляет пакет с `sender_id = 0`, сервер заменяет его на ID текущего WebSocket-соединения.

## Сценарии

### Подключение и авторизация

1. Клиент подключается к `/velora` по WebSocket.
2. Сервер повышает HTTP-соединение до WebSocket и регистрирует клиента в хабе.
3. Клиент получает `IdMessage` со своим серверным ID.
4. Клиент отправляет `LoginRequestMessage` или `RegisterRequestMessage`.
5. Сервер валидирует запрос, обращается к PostgreSQL и отвечает `OkResponseMessage` или `DenyResponseMessage`.
6. После успешного входа или регистрации клиент переводится в state `Authenticated`.

### Комната и матч

1. Авторизованный клиент отправляет `CreateRoomRequestMessage`.
2. Сервер создает комнату, назначает создателя владельцем и отправляет `RoomStateSnapshotMessage`.
3. Другие авторизованные клиенты входят через `JoinRoomRequestMessage`.
4. При join, leave и ready сервер рассылает участникам комнаты свежий `RoomStateSnapshotMessage`.
5. Игроки меняют готовность через `ReadyRequestMessage`.
6. Владелец отправляет `StartGameRequestMessage`.
7. Если комната в статусе `ROOM_STATUS_WAITING` и все игроки готовы, сервер переводит комнату в `ROOM_STATUS_STARTED` и отправляет `MatchStartMessage`.

### Список комнат

`RoomListRequestMessage` предназначен для получения краткого списка комнат. Сервер отвечает `RoomListSnapshotMessage`, где каждая комната содержит `name`, `roomId`, `playersCount`, `maxPlayer` и `status`. Список игроков в summary не передается; подробный состав комнаты приходит через `RoomStateSnapshotMessage` для конкретной комнаты.

## Разработка

Проверка сборки и тестов:

```bash
go test ./...
```

Если системный Go cache недоступен для записи, используйте кеш в `/tmp`:

```bash
env GOCACHE=/tmp/go-build-cache go test ./...
```

Перегенерация Go-кода после изменения protobuf-схемы:

```bash
make proto
```

`make proto` запускает `protoc` с явным путем к `protoc-gen-go`, поэтому команда не зависит от `PATH` IDE.

## Текущие ограничения

- Lobby-состояние хранится только в памяти и теряется при рестарте сервера.
- `RoomListSnapshotMessage` содержит только краткое summary комнаты без списка игроков.
- Нет миграций базы данных.
- Нет проверки origin при WebSocket upgrade: `CheckOrigin` сейчас разрешает любые источники.
- Нет TLS, rate limiting и отдельного слоя валидации размера/частоты сообщений.
- Нет отдельного matchmaking/gameplay слоя после `MatchStartMessage`.

## Лицензия

См. [`LICENSE`](./LICENSE).
