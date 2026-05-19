# Подробный план добавления lobby-логики

## Кратко

Добавить серверную lobby-логику поверх текущего WebSocket/protobuf-протокола.

Клиентский flow:

1. Клиент подключается по WebSocket.
2. Сервер выдает `client_id`.
3. Клиент проходит `login` или `register`.
4. После авторизации клиент может:
   - создать комнату;
   - войти в комнату;
   - выйти из комнаты;
   - переключить `ready/unready`;
   - получать snapshot состояния комнаты;
   - получить события `match_starting` и `match_started`.

Комнаты в первой версии хранятся только в памяти сервера. PostgreSQL используется только для авторизации пользователей.

## Протокол

Обновить `shared/packets.proto` и затем перегенерировать Go-код:

```bash
make proto
```

Добавить enum:

```proto
enum RoomStatus {
  ROOM_STATUS_WAITING = 0;
  ROOM_STATUS_STARTING = 1;
  ROOM_STATUS_STARTED = 2;
}
```

Добавить request-сообщения:

```proto
message CreateRoomRequestMessage {
  uint32 max_players = 1;
}

message JoinRoomRequestMessage {
  uint64 room_id = 1;
}

message LeaveRoomRequestMessage {}

message ReadyRequestMessage {
  bool ready = 1;
}
```

Добавить server-сообщения:

```proto
message RoomPlayerMessage {
  uint64 user_id = 1;
  uint64 client_id = 2;
  string username = 3;
  bool ready = 4;
  bool owner = 5;
}

message RoomStateSnapshotMessage {
  uint64 room_id = 1;
  uint32 max_players = 2;
  RoomStatus status = 3;
  repeated RoomPlayerMessage players = 4;
}

message MatchStartingMessage {
  uint64 room_id = 1;
  int64 starts_at_unix_ms = 2;
}

message MatchStartedMessage {
  uint64 room_id = 1;
  uint64 match_id = 2;
}
```

Добавить новые типы в `Packet.oneof`:

```proto
CreateRoomRequestMessage create_room_request = 8;
JoinRoomRequestMessage join_room_request = 9;
LeaveRoomRequestMessage leave_room_request = 10;
ReadyRequestMessage ready_request = 11;
RoomStateSnapshotMessage room_state_snapshot = 12;
MatchStartingMessage match_starting = 13;
MatchStartedMessage match_started = 14;
```

Добавить helper-функции в `server/pkg/packets/utils.go`:

- `NewCreateRoomRequest`
- `NewJoinRoomRequest`
- `NewLeaveRoomRequest`
- `NewReadyRequest`
- `NewRoomStateSnapshot`
- `NewMatchStarting`
- `NewMatchStarted`

## Авторизация и состояния клиента

Расширить `ClientInterface`, чтобы state-слой мог работать с авторизованным пользователем:

```go
SetUser(user *db.User)
User() *db.User
IsAuthenticated() bool
Hub() *Hub
```

Если не хочется отдавать state-слою весь `Hub`, вместо `Hub()` можно добавить более узкий метод доступа к lobby.

В `WebSocketClient` добавить поле:

```go
user *db.User
```

После успешного `login`:

- сохранить найденного пользователя через `SetUser(user)`;
- отправить `OkResponseMessage`;
- перевести клиента в новое состояние `Authenticated`.

После успешного `register`:

- использовать пользователя, возвращенного `CreateUser`;
- сохранить его через `SetUser(user)`;
- отправить `OkResponseMessage`;
- перевести клиента в `Authenticated`.

Добавить состояние, например:

```text
server/Internal/server/states/authenticated.go
```

`Authenticated` должен обрабатывать lobby request-сообщения:

- `CreateRoomRequestMessage`;
- `JoinRoomRequestMessage`;
- `LeaveRoomRequestMessage`;
- `ReadyRequestMessage`.

Если нужно сохранить текущую тестовую broadcast-логику, `Authenticated` также может продолжить обрабатывать `ChatMessage`.

Неавторизованное состояние `Connection` должно продолжать обрабатывать только:

- `LoginRequestMessage`;
- `RegisterRequestMessage`;
- `ChatMessage`, если он остается для отладки.

Если в `Connection` приходит lobby-команда, сервер отвечает:

```text
DenyResponseMessage("authentication required")
```

## Lobby manager

Добавить lobby-логику, например в файл:

```text
server/Internal/server/lobby.go
```

В `Hub` добавить поле:

```go
Lobby *LobbyManager
```

Инициализировать `LobbyManager` в `NewHub()`.

Внутренние структуры:

```go
type LobbyManager struct {
    mu sync.Mutex
    rooms map[uint64]*Room
    userRoom map[uint64]uint64
    roomIDGenerator *Internal.IdGenerator
    matchIDGenerator *Internal.IdGenerator
}

type Room struct {
    ID uint64
    MaxPlayers uint32
    Status RoomStatus
    Players map[uint64]*RoomPlayer
    PlayerOrder []uint64
    StartCancel context.CancelFunc
}

type RoomPlayer struct {
    UserID uint64
    ClientID uint64
    Username string
    Ready bool
    Owner bool
    Client server.ClientInterface
}
```

Назначение полей:

- `rooms` хранит комнаты по `room_id`;
- `userRoom` быстро проверяет, находится ли пользователь уже в комнате;
- `PlayerOrder` нужен для стабильного snapshot и передачи owner самому раннему оставшемуся игроку;
- `StartCancel` нужен для отмены countdown перед стартом матча.

## Основные методы lobby

Добавить методы:

```go
CreateRoom(client ClientInterface, maxPlayers uint32) error
JoinRoom(client ClientInterface, roomID uint64) error
LeaveRoom(client ClientInterface) error
SetReady(client ClientInterface, ready bool) error
RemoveClient(client ClientInterface) error
```

### CreateRoom

Поведение:

- проверить, что клиент авторизован;
- проверить, что пользователь еще не находится в комнате;
- если `maxPlayers == 0`, использовать `2`;
- разрешенный диапазон `maxPlayers`: `2..8`;
- создать комнату со статусом `WAITING`;
- добавить создателя как owner;
- выставить создателю `ready = false`;
- сохранить связь `userRoom[userID] = roomID`;
- отправить snapshot создателю.

Ошибки:

- `authentication required`;
- `already in room`;
- `invalid max players`.

### JoinRoom

Поведение:

- проверить авторизацию;
- проверить, что пользователь еще не находится в комнате;
- найти комнату;
- отказать, если статус не `WAITING`;
- отказать, если комната заполнена;
- добавить игрока с `ready = false`;
- сохранить связь `userRoom[userID] = roomID`;
- разослать snapshot всем участникам комнаты;
- проверить, не нужно ли стартовать матч.

Ошибки:

- `authentication required`;
- `already in room`;
- `room not found`;
- `room is full`;
- `room is not joinable`.

### LeaveRoom

Поведение:

- проверить, что пользователь находится в комнате;
- удалить игрока из комнаты;
- удалить `userRoom[userID]`;
- если комната стала пустой, удалить комнату;
- если вышел owner, назначить owner первого игрока из `PlayerOrder`;
- если комната была `STARTING`, отменить countdown и вернуть статус `WAITING`;
- разослать snapshot оставшимся участникам.

Ошибки:

- `authentication required`;
- `not in room`.

### SetReady

Поведение:

- проверить, что пользователь находится в комнате;
- отказать, если статус комнаты не `WAITING`;
- обновить `Ready`;
- разослать snapshot всем участникам;
- проверить условие старта матча.

Ошибки:

- `authentication required`;
- `not in room`;
- `room is starting`;
- `room already started`.

### RemoveClient

Поведение:

- использовать при disconnect;
- если клиент авторизован и находится в комнате, выполнить ту же cleanup-логику, что и `LeaveRoom`;
- не отправлять deny отключенному клиенту;
- snapshot отправлять только оставшимся участникам;
- повторный вызов должен быть безопасным.

## Snapshot и рассылка

Сделать внутренний метод:

```go
buildSnapshot(room *Room) packets.Msg
```

Snapshot должен включать:

- `room_id`;
- `max_players`;
- `status`;
- список игроков в порядке `PlayerOrder`;
- для каждого игрока:
  - `user_id`;
  - `client_id`;
  - `username`;
  - `ready`;
  - `owner`.

Сделать внутренний метод:

```go
broadcastToRoom(room *Room, msg packets.Msg)
```

Правило конкурентности: не держать mutex во время отправки в socket channel. Под mutex собрать snapshot и список клиентов, затем отпустить lock и выполнить отправку.

## Match starting и match started

Условие старта:

- статус комнаты `WAITING`;
- количество игроков равно `MaxPlayers`;
- все игроки имеют `Ready = true`.

Когда условие выполнено:

1. Перевести комнату в `STARTING`.
2. Создать `context.WithCancel`.
3. Сохранить cancel в комнате.
4. Вычислить `starts_at_unix_ms = now + 3 seconds`.
5. Разослать `RoomStateSnapshotMessage`.
6. Разослать `MatchStartingMessage`.

Запустить goroutine countdown:

1. Ждать 3 секунды или cancel.
2. Если пришел cancel, ничего не делать.
3. После таймера снова взять lock.
4. Проверить, что комната существует.
5. Проверить, что статус все еще `STARTING`.
6. Проверить, что игроков все еще `MaxPlayers`.
7. Проверить, что все игроки все еще ready.
8. Перевести комнату в `STARTED`.
9. Сгенерировать `match_id`.
10. Разослать `RoomStateSnapshotMessage`.
11. Разослать `MatchStartedMessage`.

Если игрок выходит во время `STARTING`:

- вызвать cancel;
- очистить `StartCancel`;
- вернуть статус `WAITING`;
- разослать новый snapshot.

В первой версии после `STARTED` комната остается в памяти со статусом `STARTED`. Дальнейшая игровая логика будет добавлена отдельно.

## Интеграция с WebSocket close

В `WebSocketClient.Close(reason string)` перед удалением клиента из hub вызвать lobby cleanup:

```go
c.hub.Lobby.RemoveClient(c)
```

Нужно учитывать, что `Close` может быть вызван и из read pump, и из write pump. Cleanup должен быть идемпотентным: повторный вызов не должен ломать состояние и не должен дважды рассылать некорректные события.

## Ошибки и ответы клиенту

Для ошибок использовать существующий `DenyResponseMessage`.

Рекомендуемые причины:

- `authentication required`;
- `already in room`;
- `room not found`;
- `room is full`;
- `room is not joinable`;
- `not in room`;
- `invalid max players`;
- `room is starting`;
- `room already started`.

Для успешных lobby-команд не отправлять `OkResponseMessage`, потому что успешный результат представлен snapshot/event-сообщениями.

## Тесты

Добавить unit-тесты lobby manager. Для тестов лучше использовать fake client, реализующий нужную часть `ClientInterface` и сохраняющий отправленные сообщения в slice.

Покрыть сценарии:

- `CreateRoom` с `max_players = 0` создает комнату на 2 игрока;
- `CreateRoom` с `max_players = 4` создает комнату на 4 игрока;
- `CreateRoom` отклоняет `max_players < 2`;
- `CreateRoom` отклоняет `max_players > 8`;
- пользователь не может создать вторую комнату, если уже в комнате;
- `JoinRoom` добавляет игрока и рассылает snapshot;
- `JoinRoom` отклоняет несуществующую комнату;
- `JoinRoom` отклоняет заполненную комнату;
- `JoinRoom` отклоняет комнату в `STARTING`;
- `JoinRoom` отклоняет комнату в `STARTED`;
- `LeaveRoom` удаляет игрока и обновляет snapshot;
- пустая комната удаляется;
- при выходе owner ownership переходит следующему игроку;
- `SetReady(true)` обновляет ready-состояние;
- `SetReady(false)` снимает ready-состояние;
- матч не стартует, если комната не заполнена;
- матч не стартует, если не все игроки ready;
- матч переходит в `STARTING`, когда комната заполнена и все ready;
- после countdown матч переходит в `STARTED`;
- выход игрока во время countdown отменяет старт и возвращает комнату в `WAITING`;
- `RemoveClient` удаляет игрока из комнаты;
- повторный `RemoveClient` безопасен.

Проверка всей базы:

```bash
go test ./...
```

## Допущения

- Lobby-состояние хранится только в памяти и теряется при рестарте сервера.
- Один WebSocket-клиент соответствует одному авторизованному пользователю.
- Повторный `login/register` на уже авторизованном клиенте не нужен; можно отвечать deny.
- `register` после успешного создания пользователя сразу авторизует клиента.
- `room_id` и `match_id` генерируются сервером как `uint64`.
- Максимальный размер комнаты в первой версии ограничен `8`.
- Countdown перед стартом матча равен 3 секундам.
- Успешные lobby-команды подтверждаются snapshot/event-сообщениями, а не `OkResponseMessage`.
