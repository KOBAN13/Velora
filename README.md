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
- In-memory match runtime с фиксированным tick rate, фазами матча и рассылкой `MatchSnapshotMessage`.
- ECS-слой для игровых сущностей: player cells, cores, nutrients и walls.
- System runner для фаз, input, movement, nutrient spawn, wall gate и cleanup-логики.
- Загрузка стартовых игровых параметров из Google Sheets.
- Protobuf snapshots для состояния комнаты, краткого списка комнат и состояния матча.
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
├── esc                                  # ECS world, components, resources, commands и queries
├── systems                              # gameplay systems и system runner
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
    │       ├── match                    # lifecycle матча, tick loop и match snapshots
    │       ├── spawners                 # стартовые позиции и ресурсы спавна
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

## Обновление ECS-зависимости

Проект использует модуль `github.com/KOBAN13/kukuruzka-esc`, пакет `github.com/KOBAN13/kukuruzka-esc/ecs`.

Для локальной разработки зависимости без commit/push:

```bash
make ecs-local KUKURUZKA_ESC_PATH=../kukuruzka-esc
```

В Windows CMD без `make`:

```cmd
scripts\kukuruzka-esc.cmd local ..\kukuruzka-esc
```

Эта команда добавляет в `go.mod` `replace` на локальную копию и выполняет `go mod tidy`.

Чтобы убрать локальный `replace` и подтянуть версию из GitHub:

```bash
make ecs-update
```

В Windows CMD:

```cmd
scripts\kukuruzka-esc.cmd update
```

Можно указать конкретную ветку, тег или коммит:

```bash
make ecs-update KUKURUZKA_ESC_REF=main
make ecs-update KUKURUZKA_ESC_REF=v0.1.0
```

В Windows CMD:

```cmd
scripts\kukuruzka-esc.cmd update main
scripts\kukuruzka-esc.cmd update v0.1.0
```

Чтобы запускать обновление по таймеру раз в 10 секунд:

```bash
make ecs-watch-update KUKURUZKA_ESC_REF=main
```

В Windows CMD:

```cmd
scripts\kukuruzka-esc.cmd watch-update main 10
```

Интервал можно поменять:

```bash
make ecs-watch-update KUKURUZKA_ESC_REF=main KUKURUZKA_ESC_INTERVAL=30
```

В Windows CMD:

```cmd
scripts\kukuruzka-esc.cmd watch-update main 30
```

Команда работает до остановки через `Ctrl+C`. Она не делает commit/push, только повторяет обновление зависимости из Git.

Чтобы закоммитить и запушить изменения в локальном репозитории зависимости, а затем обновить `Velora`:

```bash
make ecs-publish KUKURUZKA_ESC_PATH=../kukuruzka-esc KUKURUZKA_ESC_MESSAGE="update ecs"
```

В Windows CMD:

```cmd
scripts\kukuruzka-esc.cmd publish ..\kukuruzka-esc "update ecs"
```

Если нужно обновиться не на текущую ветку зависимости, а на конкретный ref:

```bash
make ecs-publish KUKURUZKA_ESC_PATH=../kukuruzka-esc KUKURUZKA_ESC_MESSAGE="update ecs" KUKURUZKA_ESC_PUBLISH_REF=v0.1.0
```

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
    PlayerInputMessage player_input = 18;
    MatchSnapshotMessage match_snapshot = 19;
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

Match сообщения:

- `MatchStartMessage` - уведомление о старте матча: `roomId`, `matchId`, `player_id`, `slot`, `mapSeed`, `startsAtUnixMs`.
- `PlayerInputMessage` - ввод игрока для матча: `matchId` и направление движения `movePosition`.
- `MatchSnapshotMessage` - серверный snapshot матча: `matchId`, `serverTick`, `phase`, `phaseTimeLeftMs`, `playerCells`, `cores`, `nutrients`, `walls`.
- `MatchPhase` - фаза матча: `MATCH_PHASE_PREPARE`, `MATCH_PHASE_ACTIVE`, `MATCH_PHASE_ENDED`.

Snapshots и ответы:

- `RoomStateSnapshotMessage` - состояние одной комнаты: `roomId`, `maxPlayer`, `status`, список игроков.
- `RoomPlayerMessage` - игрок комнаты: `userId`, `clientId`, `username`, `isReady`, `owner`.
- `RoomListSnapshotMessage` - краткий список комнат.
- `RoomSummaryMessage` - summary комнаты: `name`, `roomId`, `Players`, `maxPlayer`, `status`.
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
8. Матч инициализирует ECS systems, стартовые сущности и начальные nutrients.
9. После старта `Match.Run` запускает tick loop с частотой `20` тиков в секунду.
10. На каждом тике сервер синхронизирует inputs в ECS resources, выполняет systems по стадиям, применяет command buffer и рассылает `MatchSnapshotMessage` подключенным игрокам.

### Match tick flow

`Match.Tick` работает под mutex матча, чтобы inputs, ECS world и список подключенных клиентов оставались согласованными в рамках одного тика.

Порядок обработки:

1. Увеличить `ServerTick`.
2. Скопировать актуальные inputs игроков в `esc.Resources`.
3. Создать `esc.SystemContext` с tick, delta time, временем, фазой, command buffer и resources.
4. Выполнить `SystemRunner.UpdateSystems` по стадиям: phase, input, movement, spawn, rules, cleanup.
5. После каждой стадии применить накопленные ECS commands к world.
6. Сохранить обновленные `Phase` и `PhaseEndsAt` обратно в `Match`.
7. Собрать и разослать `MatchSnapshotMessage`.

Игровые системы сейчас отвечают за смену фаз, перенос player input в direction, движение активных player cells, подбор и спавн nutrients, открытие walls после prepare-фазы и деактивацию умерших player cells.

### Список комнат

`RoomListRequestMessage` предназначен для получения списка комнат. Сервер отвечает `RoomListSnapshotMessage`, где каждая комната содержит `name`, `roomId`, `Players`, `maxPlayer` и `status`. Подробный состав конкретной комнаты также приходит через `RoomStateSnapshotMessage`.

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
- Match runtime хранится только в памяти и останавливается при потере всех подключенных клиентов.
- Нет механизма восстановления клиента в уже запущенный матч после реконнекта.

## Лицензия

См. [`LICENSE`](./LICENSE).
