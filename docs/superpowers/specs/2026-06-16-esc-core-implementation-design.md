# ESC Core Implementation Design

Дата: 2026-06-16

Статус: согласованный дизайн для реализации v1.

## 1. Контекст

В проекте уже есть пакет `esc`, но его текущая модель ближе к typed entity structs:

```text
World.Entities map[EntityId]Entity
PlayerCell/Core/Nutrient/Wall as concrete structs
QueryPlayerCells/QueryActiveNutrients scan all entities
```

Эта модель работает для первых систем, но плохо масштабируется:

- каждый query проходит весь `World.Entities`;
- каждый query аллоцирует новый slice;
- вложенные системы вроде pickup делают repeated full scans;
- entity kind и набор данных смешаны в одном struct;
- текущий API сложно развивать до reusable ECS framework.

Цель новой версии: построить отдельный универсальный ECS core, который можно использовать в Velora и показать как самостоятельную инженерную часть проекта.

Новый пакет называется `esc_core`. Текущий `esc` пока остается в репозитории, чтобы миграция была поэтапной.

## 2. Цели v1

V1 должна быть не игрушечной, но и без лишнего polishing. Обязательные свойства:

- strict component-based ECS;
- generational entity handles;
- archetype storage;
- Struct of Arrays внутри archetype tables;
- read/write query access;
- query создаются при setup системы, а не каждый `Update`;
- lazy refresh query cache при появлении новых archetype;
- command buffer для structural changes во время update;
- прямые value writes через query/world;
- bundles как часть framework;
- resources как singleton state;
- system runner с game-level stages;
- validation read/write конфликтов внутри stage;
- debug reports по systems, access, queries, archetypes;
- все ошибки возвращаются как `error`, без `panic`.

## 3. Не цели v1

Эти вещи не входят в первую версию:

- parallel scheduler;
- events/signals;
- serialization/snapshot framework внутри `esc_core`;
- rollback/replay;
- prefab/editor tooling;
- relations/parent-child;
- change detection;
- enabled components/disabled masks;
- spatial grid/quadtree внутри ECS core;
- code generation.

Spatial index позже может быть отдельным resource/system поверх ECS. Он не должен смешиваться с archetype storage.

## 4. Пакеты и ответственность

### `esc_core`

Reusable ECS framework.

Отвечает за:

- entity lifecycle;
- component type registry;
- archetype tables;
- query descriptors и iterators;
- command buffer;
- resources;
- system runner contracts;
- stage execution mechanics;
- access validation;
- debug/inspection reports.

Не знает про:

- PlayerCell;
- Nutrient;
- protobuf;
- match phases Velora;
- websocket/client state;
- server config.

### `esc`

Текущий пакет старой ECS-модели. На время миграции остается как compatibility layer.

После завершения миграции возможны варианты:

- удалить `esc`;
- оставить только игровые компоненты и bundles;
- переименовать игровые компоненты в отдельный пакет, например `game_components`.

Решение принимается после миграции систем.

### `systems`

Gameplay systems Velora.

Отвечают за:

- input;
- movement;
- nutrient pickup/spawn;
- wall gate logic;
- death/cleanup;
- phase changes.

Системы зависят от `esc_core` и игровых компонентов. Они не должны знать внутренности archetype storage.

### `match`

Authoritative server runtime.

Отвечает за:

- tick loop;
- locks;
- player refs;
- match metadata;
- system context creation;
- command application boundaries;
- snapshot sending.

`match` владеет `World`, `Resources` и `Runner`.

## 5. Главные архитектурные решения

### 5.1 Strict component-based ECS

Entity больше не хранит данные и не имеет concrete struct вроде `PlayerCell`.

Entity - это handle. Игровые объекты описываются только набором компонентов.

```text
PlayerCell =
  PlayerTag
  Owner
  Position
  MoveDirection
  Health
  MaxHealth
  Biomass
  Level
  Active

Core =
  CoreTag
  Owner
  Position
  Health
  MaxHealth

Nutrient =
  NutrientTag
  Position
  NutrientValue
  Active

Wall =
  WallTag
  WallState
```

Tags являются обычными zero-sized components:

```go
type PlayerTag struct{}
type CoreTag struct{}
type NutrientTag struct{}
type WallTag struct{}
```

### 5.2 Bool-состояния остаются value components

`Active`, `Visible`, `Dead`, `Invulnerable` и похожие состояния не должны превращаться в add/remove tags в v1.

Правило:

```text
Компонент добавляется/удаляется, когда меняется природа entity.
Поле компонента мутируется, когда меняется состояние entity.
```

Пример:

```go
type Active struct {
	Value bool
}
```

Плюс: переключение дешево и не вызывает перенос entity между archetypes.

Минус: query по `NutrientTag + Active` проходит и активные, и неактивные nutrients, а система фильтрует `active.Value`. Для v1 это приемлемо.

### 5.3 Value writes сразу, structural writes через commands

Value mutation применяется сразу:

```go
pos, err := esc_core.Write[Position](it)
if err != nil {
	return err
}

pos.X += dx
pos.Y += dy
```

Structural mutation во время system update идет через `CommandBuffer`:

```go
err := ctx.Commands.Spawn().
	Bundle(NewNutrientBundle(position, value)).
	Commit()
```

Structural операции:

- spawn entity;
- despawn entity;
- add component;
- remove component.

Direct structural operations на `World` разрешены вне active update:

- setup;
- tests;
- initialization;
- migration code;
- command buffer apply phase.

Во время active update direct structural calls вроде `esc_core.Spawn(world, ...)`, `esc_core.Add(world, ...)`, `esc_core.Remove(world, ...)`, `esc_core.Despawn(world, ...)` должны возвращать error.

### 5.4 Stage boundary command application

Сохраняем текущую модель Velora:

```text
Stage starts
  systems run sequentially
  systems may enqueue structural commands
Stage ends
  command buffer applies
Next stage sees updated world
```

Внутри stage мир структурно стабилен. Это защищает query iterators от invalidation.

## 6. Go API constraint

Текущая версия Go не поддерживает generic methods:

```go
// This does not compile:
func (b *QueryBuilder) Read[T any]() {}
```

Поэтому все typed операции, которым нужен `[T]`, должны быть package-level generic functions.

Вместо:

```go
esc_core.Query(world).
	Write[Position]().
	Read[Active]()
```

используем:

```go
builder := esc_core.NewQuery(world, "players")
esc_core.QueryWith[PlayerTag](builder)
esc_core.QueryWrite[Position](builder)
esc_core.QueryRead[MoveDirection](builder)
esc_core.QueryRead[Active](builder)

players, err := builder.Build()
```

Для iterator:

```go
active, err := esc_core.Read[Active](it)
pos, err := esc_core.Write[Position](it)
```

Для component type tokens:

```go
component := esc_core.ComponentType[Poisoned]()
```

Это менее красиво, чем generic methods, но компилируемо и сохраняет type-safety на границе API.

## 7. Entity model

Entity - generational handle.

```go
type Entity struct {
	index      uint32
	generation uint32
}
```

Публичный API может скрывать поля:

```go
func (e Entity) Index() uint32
func (e Entity) Generation() uint32
func (e Entity) IsZero() bool
func (e Entity) String() string
```

`World` хранит:

```go
type entitySlot struct {
	generation uint32
	alive      bool
	location   entityLocation
}

type entityLocation struct {
	archetype *archetype
	row       int
}
```

Также нужен freelist:

```go
freeEntityIndexes []uint32
```

Spawn:

- берет index из freelist или создает новый;
- использует текущую generation;
- создает/находит archetype по component signature;
- вставляет entity в archetype row;
- сохраняет location.

Despawn:

- проверяет generation;
- удаляет row из archetype через swap-remove;
- обновляет location entity, которую swap переместил;
- помечает slot как dead;
- increment generation;
- возвращает index во freelist.

Преимущество: старый handle не может случайно обратиться к новой entity после reuse index.

## 8. Component registry

`esc_core` должен уметь назначать каждому component type стабильный `ComponentID` внутри процесса.

```go
type ComponentID uint32

type ComponentInfo struct {
	ID   ComponentID
	Name string
	Type reflect.Type
	Size uintptr
	IsTag bool
}
```

Регистрация происходит lazy через generic function:

```go
id := esc_core.ComponentIDOf[Position](world)
```

или через helper:

```go
typ := esc_core.ComponentType[Position]()
```

Важно: component registry принадлежит `World` или `Registry`, а не global singleton. Это делает tests независимыми.

Решение v1:

```text
World owns ComponentRegistry.
ComponentID стабилен внутри одного World.
```

Компонент должен быть struct type. Pointer components запрещены в v1, чтобы не смешивать ownership.

Допустимо:

```go
type Position struct { X, Y float32 }
type PlayerTag struct{}
```

Не допустимо:

```go
*Position
map[string]int
[]Item
interface{}
```

Это ограничение можно смягчить позже, но v1 лучше держать простой и data-oriented.

## 9. Archetype storage

Archetype - таблица entity с одинаковым набором компонентов.

```go
type archetype struct {
	id        archetypeID
	signature componentSignature
	entities  []Entity
	columns   map[ComponentID]column
}
```

`componentSignature` должен позволять:

- сравнить точное равенство наборов компонентов;
- проверить contains all;
- проверить excludes none;
- получить stable key для map.

Для v1 можно реализовать signature как sorted `[]ComponentID` plus string/key cache:

```go
type componentSignature struct {
	ids []ComponentID
	key string
}
```

Позже можно заменить на bitset, если ComponentID станет компактным и потребуется больше скорости.

### 9.1 Columns

Column хранит значения одного component type для всех rows archetype.

Внутри нужен typed storage:

```go
type typedColumn[T any] struct {
	values []T
}
```

Снаружи archetype держит type-erased interface:

```go
type column interface {
	Len() int
	AppendZero()
	AppendAny(value any) error
	SwapRemove(row int)
	CopyValueTo(row int, target column) error
}
```

Для hot path query не должен делать reflection на каждую entity.

Compiled query должен один раз на archetype получить typed column pointers:

```go
posColumn := column.(*typedColumn[Position])
```

Iterator после этого ходит по slices напрямую.

### 9.2 Row removal

Удаление row делается через swap-remove:

```text
remove row i
last row moves into i
all columns swap-remove same row
entities swap-remove same row
World updates moved entity location
```

Это сохраняет плотные arrays и O(1) despawn/remove.

Порядок entity внутри archetype не гарантируется. Snapshot builder, если нужен deterministic output, сортирует результат по entity handle/id на уровне game snapshot.

## 10. World API

Основные функции:

```go
func NewWorld(options ...WorldOption) *World

func Spawn(world *World, components ...any) (Entity, error)
func Despawn(world *World, entity Entity) error

func Add(world *World, entity Entity, components ...any) error
func Remove(world *World, entity Entity, componentTypes ...ComponentType) error

func Has[T any](world *World, entity Entity) (bool, error)
func Get[T any](world *World, entity Entity) (T, bool, error)
func GetWrite[T any](world *World, entity Entity) (*T, bool, error)
func Set[T any](world *World, entity Entity, value T) error

func IsAlive(world *World, entity Entity) bool
func EntityCount(world *World) int
func ArchetypeCount(world *World) int
```

Because Go has no generic methods, all typed operations are package-level functions.

Direct structural mutation guard:

```go
type MutationPhase uint8

const (
	MutationIdle MutationPhase = iota
	MutationRunningSystem
	MutationApplyingCommands
)
```

`Spawn/Add/Remove/Despawn` return error if called while `MutationRunningSystem`.

`CommandBuffer.Apply` temporarily moves world to `MutationApplyingCommands`.

## 11. Query API

### 11.1 Builder

Query создается один раз при setup системы.

```go
builder := esc_core.NewQuery(world, "players")
esc_core.QueryWith[PlayerTag](builder)
esc_core.QueryWrite[Position](builder)
esc_core.QueryRead[MoveDirection](builder)
esc_core.QueryRead[Active](builder)

players, err := builder.Build()
```

Builder накапливает:

```go
type QueryDescriptor struct {
	Name     string
	With     []ComponentID
	Without  []ComponentID
	Reads    []ComponentID
	Writes   []ComponentID
}
```

Rules:

- `Read[T]` also implies component presence.
- `Write[T]` also implies component presence.
- `With[T]` requires component presence but does not count as read/write access.
- `Without[T]` excludes archetypes containing component.
- Same component cannot be both read and write in one query.
- Duplicate entries return error at `Build`.

Recommended tag usage:

```go
esc_core.QueryWith[PlayerTag](builder)
```

not:

```go
esc_core.QueryRead[PlayerTag](builder)
```

Tags are filters, not data dependencies.

### 11.2 Compiled query

`Build` returns compiled query:

```go
type Query struct {
	world        *World
	name         string
	descriptor   QueryDescriptor
	matches      []queryArchetypePlan
	seenVersion  uint64
	access       AccessSet
}
```

`matches` is cached list of matching archetypes.

World has:

```go
archetypeVersion uint64
```

When a new archetype is created, version increments.

At `Query.Iter()`:

- if `query.seenVersion == world.archetypeVersion`, no refresh;
- otherwise query scans only archetypes created since previous version or does full refresh in v1;
- query updates `matches` and `seenVersion`.

Full refresh is acceptable for v1 because new archetypes are rare compared to ticks.

### 11.3 Iterator

Usage:

```go
it := players.Iter()
for it.Next() {
	active, err := esc_core.Read[Active](it)
	if err != nil {
		return err
	}
	if !active.Value {
		continue
	}

	dir, err := esc_core.Read[MoveDirection](it)
	if err != nil {
		return err
	}

	pos, err := esc_core.Write[Position](it)
	if err != nil {
		return err
	}

	pos.X += dir.X * speed * ctx.DeltaSeconds
	pos.Y += dir.Y * speed * ctx.DeltaSeconds
}
```

Iterator responsibilities:

- walk cached archetype plans;
- expose current entity;
- expose read values by copy;
- expose write values by pointer;
- validate that requested component was declared with matching access.

API:

```go
func (q *Query) Iter() Iterator
func (it *Iterator) Next() bool
func (it *Iterator) Entity() Entity

func Read[T any](it *Iterator) (T, error)
func Write[T any](it *Iterator) (*T, error)
```

`Read[T]` returns a copy. `Write[T]` returns pointer.

This is intentional:

- read-only components cannot be mutated accidentally through returned pointer;
- write access is explicit in query descriptor and access validation.

## 12. Access validation

Each query exposes access metadata:

```go
type AccessSet struct {
	Reads  ComponentSet
	Writes ComponentSet
}
```

Conflict rules:

```text
Read + Read = ok
Read + Write = conflict
Write + Write = conflict
```

`With` and `Without` do not count as data access.

Structural commands are deferred to stage boundary, so they do not conflict with query iteration inside the same stage. Direct structural world mutation is blocked during system update.

Runner validation:

```go
func (r *Runner) ValidateAccess() error
func (r *Runner) DebugAccess() string
```

Validation runs per stage. If v1 runner executes systems sequentially, conflicts are still useful:

- warn/error before future parallel scheduler;
- catch accidental multiple writers;
- document system dependencies.

Decision for v1:

```text
ValidateAccess returns error.
Application decides whether to log and continue or fail runner build.
```

No panic.

## 13. Systems and runner

### 13.1 System contract

`esc_core` should define minimal generic contracts:

```go
type System interface {
	Name() string
	Update(ctx *Context) error
	Access() AccessSet
	DebugQueries() []QueryDebugInfo
}
```

Stage itself should be generic enough for game-level stages.

Possible contract:

```go
type StageID uint16

type StagedSystem interface {
	System
	Stage() StageID
}
```

Velora can define:

```go
const (
	StagePhase esc_core.StageID = iota + 1
	StageInput
	StageMovement
	StageSpawn
	StageRules
	StageCleanup
)
```

This keeps stages in game layer, but runner can order and execute them.

### 13.2 Query ownership

Systems create queries during construction/setup:

```go
type MovementSystem struct {
	players *esc_core.Query
}

func NewMovementSystem(world *esc_core.World) (*MovementSystem, error) {
	builder := esc_core.NewQuery(world, "players")
	esc_core.QueryWith[PlayerTag](builder)
	esc_core.QueryWrite[Position](builder)
	esc_core.QueryRead[MoveDirection](builder)
	esc_core.QueryRead[Active](builder)

	players, err := builder.Build()
	if err != nil {
		return nil, err
	}

	return &MovementSystem{players: players}, nil
}
```

`Update` only iterates:

```go
func (s *MovementSystem) Update(ctx *esc_core.Context) error {
	it := s.players.Iter()
	for it.Next() {
		// read/write components
	}
	return nil
}
```

No query allocation in hot update loop.

### 13.3 Access collection

System implementation can expose access manually:

```go
func (s *MovementSystem) Access() esc_core.AccessSet {
	return esc_core.MergeAccess(s.players.Access())
}
```

Or use helper:

```go
type QueryOwner struct {
	queries []*Query
}
```

V1 should start simple:

- system stores queries;
- `Access()` returns merged query access;
- `DebugQueries()` returns query debug info.

Avoid heavy `SystemBase` until repetition becomes painful.

## 14. Context and resources

`Context` passed to systems:

```go
type Context struct {
	World        *World
	Commands     *CommandBuffer
	Resources    *Resources
	Tick         uint64
	DeltaSeconds float32
	Stage        StageID
}
```

Resources are singleton data, not components:

- input map;
- nutrient spawner;
- RNG;
- match config;
- clock/time data if needed.

Typed resource API must also avoid generic methods.

```go
func PutResource[T any](resources *Resources, value T) error
func GetResource[T any](resources *Resources) (*T, bool, error)
func RemoveResource[T any](resources *Resources) error
```

Resources should use type registry similar to components, but separate from component registry.

## 15. Command buffer

### 15.1 Scope

V1 command buffer handles structural mutations:

- spawn;
- despawn;
- add components;
- remove components.

Value `Set` command is intentionally not included in v1.

Reason: value writes already happen immediately through `Write[T]` or `Set[T]`. Adding deferred `Set` creates two visibility models for ordinary value changes.

If rollback/replay becomes a goal, deferred value commands can be added later.

### 15.2 Builder-style API

Commands are builder-style, but without generic methods.

Spawn:

```go
err := ctx.Commands.Spawn().
	With(Position{X: 1, Y: 2}).
	With(Health{Value: 100}).
	With(PlayerTag{}).
	Commit()
```

With bundle:

```go
err := ctx.Commands.Spawn().
	Bundle(NewPlayerCellBundle(...)).
	Commit()
```

Add:

```go
err := ctx.Commands.Add(entity).
	With(Poisoned{Duration: 3}).
	Commit()
```

Remove:

```go
remove := ctx.Commands.Remove(entity)
remove.Component(esc_core.ComponentType[Poisoned]())
err := remove.Commit()
```

Despawn:

```go
err := ctx.Commands.Despawn(entity)
```

### 15.3 Command errors

All command builder methods return or store errors. `Commit()` returns error.

Rules:

```text
Spawn without components = error
Duplicate component in one command = error
Add without components = error
Add existing component = error
Remove without component types = error
Remove missing component = no-op
Despawn missing/dead entity = no-op
Use builder after Commit = error
```

Command buffer also exposes:

```go
func (c *CommandBuffer) Apply(world *World) error
func (c *CommandBuffer) Clear()
func (c *CommandBuffer) Len() int
func (c *CommandBuffer) Errors() []error
```

Application logging policy lives outside `esc_core`.

## 16. Bundles

Bundles are part of `esc_core`, but game-specific bundles live in game packages.

Framework contract:

```go
type Bundle interface {
	Apply(*BundleBuilder) error
}

type BundleFunc func(*BundleBuilder) error

func (f BundleFunc) Apply(b *BundleBuilder) error {
	return f(b)
}
```

`BundleBuilder` accepts concrete component values:

```go
func (b *BundleBuilder) With(component any) error
```

Game usage:

```go
func NewPlayerCellBundle(ownerID uint64, position Position, cfg PlayerCellConfig) esc_core.Bundle {
	return esc_core.BundleFunc(func(b *esc_core.BundleBuilder) error {
		if err := b.With(PlayerTag{}); err != nil { return err }
		if err := b.With(Owner{UserID: ownerID}); err != nil { return err }
		if err := b.With(position); err != nil { return err }
		if err := b.With(MoveDirection{}); err != nil { return err }
		if err := b.With(Health{Value: int32(cfg.HP)}); err != nil { return err }
		if err := b.With(MaxHealth{Value: int32(cfg.MaxHP)}); err != nil { return err }
		if err := b.With(Biomass{Value: uint32(cfg.Biomass)}); err != nil { return err }
		if err := b.With(Level{Value: uint32(cfg.Level)}); err != nil { return err }
		return b.With(Active{Value: cfg.Alive})
	})
}
```

Bundles are convenience only. They do not define entity types in storage.

## 17. Error model

Project decision: no panic for ECS programmer mistakes.

All invalid operations return `error`.

Examples:

- invalid entity generation;
- duplicate component;
- component type is not a struct;
- query reads component that was not declared;
- query writes component declared read-only;
- direct structural mutation during `RunningSystem`;
- access conflict validation failure;
- adding component that entity already has;
- build query with no required components.

Use typed/sentinel errors where helpful:

```go
var ErrInvalidEntity = errors.New("invalid entity")
var ErrComponentNotFound = errors.New("component not found")
var ErrDuplicateComponent = errors.New("duplicate component")
var ErrInvalidMutationPhase = errors.New("invalid mutation phase")
var ErrAccessConflict = errors.New("access conflict")
```

Prefer wrapping:

```go
return fmt.Errorf("%w: %s on %s", ErrAccessConflict, componentName, stageName)
```

This makes errors loggable and testable.

## 18. Debug and inspection

V1 must include debug reports. They can return strings or structured reports. Prefer structured core plus string formatting helpers.

World:

```go
func (w *World) InspectArchetypes() ArchetypeReport
func (w *World) DebugArchetypes() string
```

Runner:

```go
func (r *Runner) InspectAccess() AccessReport
func (r *Runner) DebugAccess() string
func (r *Runner) DebugQueries() string
func (r *Runner) ValidateAccess() error
```

Query debug belongs to systems/runner, not to `World`. Queries are owned by systems and refreshed lazily; `World` does not keep a global list of live query objects in v1.

Example `DebugArchetypes`:

```text
Archetypes: 4

#1 PlayerTag, Owner, Position, MoveDirection, Health, MaxHealth, Biomass, Level, Active
  entities: 2

#2 CoreTag, Owner, Position, Health, MaxHealth
  entities: 2

#3 NutrientTag, Position, NutrientValue, Active
  entities: 300

#4 WallTag, WallState
  entities: 1
```

Example `DebugAccess`:

```text
StageMovement
  MovementSystem
    Query players
      with: PlayerTag
      read: MoveDirection, Active
      write: Position

StageSpawn
  NutrientSystem
    Query players
      with: PlayerTag
      read: Position, Active
    Query nutrients
      with: NutrientTag
      read: Position, Active, NutrientValue
      write: Active
```

Example conflict:

```text
access conflict in StageMovement:
  component Position
  MovementSystem.players writes
  KnockbackSystem.targets writes
```

## 19. Snapshot migration

Current snapshot builder uses:

```go
world.QueryPlayerCells()
world.QueryCores()
world.QueryNutrients()
world.QueryWalls()
```

New snapshot uses `esc_core` queries:

Player cells:

```text
with: PlayerTag
read: Owner, Position, Health, Biomass, Level, Active
```

Cores:

```text
with: CoreTag
read: Owner, Position, Health
```

Nutrients:

```text
with: NutrientTag
read: Position, NutrientValue, Active
```

Walls:

```text
with: WallTag
read: WallState
```

Snapshot output should remain deterministic. Since archetype row order is not stable, snapshot builder sorts by entity before building protobuf messages.

## 20. Velora system examples

### 20.1 MovementSystem

Setup:

```go
builder := esc_core.NewQuery(world, "players")
esc_core.QueryWith[PlayerTag](builder)
esc_core.QueryWrite[Position](builder)
esc_core.QueryRead[MoveDirection](builder)
esc_core.QueryRead[Active](builder)
players, err := builder.Build()
```

Update:

```go
it := s.players.Iter()
for it.Next() {
	active, err := esc_core.Read[Active](it)
	if err != nil { return err }
	if !active.Value { continue }

	dir, err := esc_core.Read[MoveDirection](it)
	if err != nil { return err }
	if dir.IsZero() { continue }

	pos, err := esc_core.Write[Position](it)
	if err != nil { return err }

	pos.X += dir.X * DefaultPlayerSpeed * ctx.DeltaSeconds
	pos.Y += dir.Y * DefaultPlayerSpeed * ctx.DeltaSeconds
}
```

### 20.2 Nutrient pickup

Queries:

```text
players:
  with PlayerTag
  read Position, Active
  write Biomass, Level

nutrients:
  with NutrientTag
  read Position, NutrientValue
  write Active
```

For v1, nested loop remains possible:

```text
for active players:
  for active nutrients:
    if distance <= pickup:
      nutrient.Active.Value = false
      player.Biomass += nutrient.Value
      player.Level = 1 + Biomass / 100
```

This removes full-world query scans, but still has player x nutrient cost. If this becomes hot, solve it with spatial index resource, not by complicating ECS core.

## 21. Implementation phases

### Phase 1: Package skeleton and tests

Create:

```text
esc_core/
  entity.go
  component.go
  signature.go
  column.go
  archetype.go
  world.go
  query.go
  command.go
  bundle.go
  resources.go
  system.go
  runner.go
  debug.go
```

Tests start with entity lifecycle and component registration.

### Phase 2: Generational entities

Implement:

- `Entity`;
- entity slots;
- freelist;
- alive validation;
- stale generation tests.

Required tests:

- spawn creates alive entity;
- despawn invalidates entity;
- reused index gets new generation;
- old handle cannot access new entity.

### Phase 3: Component registry and signatures

Implement:

- `ComponentID`;
- `ComponentType[T]`;
- `ComponentRegistry`;
- struct-only validation;
- signature sorting/keying.

Required tests:

- same type gets same component id;
- different types get different ids;
- duplicate components are rejected;
- pointer component is rejected.

### Phase 4: Archetype tables

Implement:

- typed columns;
- append;
- swap-remove;
- copy entity between archetypes;
- location updates.

Required tests:

- spawn with same components lands in same archetype;
- spawn with different components creates different archetype;
- despawn swap-removes and updates moved entity location.

### Phase 5: World structural operations

Implement:

- `Spawn`;
- `Despawn`;
- `Add`;
- `Remove`;
- `Has/Get/GetWrite/Set`;
- mutation phase guard.

Required tests:

- add component moves entity to new archetype;
- remove component moves entity to new archetype;
- set updates in place;
- direct spawn during running system returns error.

### Phase 6: Query builder and iterator

Implement:

- query descriptor;
- package-level generic query functions;
- build validation;
- cached archetype matches;
- lazy refresh;
- iterator read/write access.

Required tests:

- query returns only matching archetypes;
- `With` filters without counting as read/write;
- `Read` returns copy;
- `Write` mutates component;
- undeclared read/write returns error;
- query sees new archetype after lazy refresh.

### Phase 7: Command buffer and bundles

Implement:

- builder commands;
- bundle interface;
- command apply;
- command error accumulation.

Required tests:

- spawn command creates entity after apply;
- add/remove commands migrate archetype after apply;
- duplicate command component returns error;
- missing despawn is no-op.

### Phase 8: Resources and context

Implement:

- typed resource storage;
- `Context`;
- resource get/put/remove.

Required tests:

- put/get resource by type;
- missing resource returns ok=false;
- wrong type cannot collide.

### Phase 9: Runner validation and debug

Implement:

- system contracts;
- stage order;
- stage-boundary command apply;
- access collection;
- `ValidateAccess`;
- debug reports.

Required tests:

- read/read systems do not conflict;
- read/write systems conflict in same stage;
- write/write systems conflict in same stage;
- same conflict in different stages is allowed;
- commands apply after stage.

### Phase 10: Velora migration

Migrate gradually:

1. Define Velora components/tags using `esc_core`.
2. Define bundles for player/core/nutrient/wall.
3. Create new world in match setup.
4. Migrate `MovementSystem`.
5. Migrate `InputSystem`.
6. Migrate `NutrientSystem`.
7. Migrate `WallGateSystem`.
8. Migrate `DeathSystem`.
9. Migrate snapshot builder.
10. Remove old `esc` usage from runtime.

Keep protobuf contract unchanged during migration.

## 22. Benchmark plan

Add benchmarks once core behavior is correct:

```text
BenchmarkSpawnSameArchetype
BenchmarkQueryIteratePositionActive
BenchmarkAddComponentMigration
BenchmarkRemoveComponentMigration
BenchmarkDespawnSwapRemove
BenchmarkVeloraMovementSystem
BenchmarkVeloraNutrientPickup
```

Compare old `esc` and new `esc_core` on:

- allocations per tick;
- query iteration time;
- movement update time;
- nutrient pickup time;
- snapshot build time.

Expected first wins:

- no full-world scan for every query;
- no query slice allocation per update;
- dense component iteration;
- clearer system access reports.

## 23. Open implementation notes

These are not product decisions; they can be adjusted during implementation if tests show a cleaner approach.

### Signature representation

Start with sorted ids + key string. Replace with bitset only if benchmarks show signature matching is hot.

### Query refresh

Start with full refresh when `world.archetypeVersion` changes. Optimize to incremental archetype append later if needed.

### Debug formatting

Keep structured report types internally. String output is convenience.

### Access validation policy

`esc_core` returns validation errors. Velora decides:

- log and continue;
- fail match/system runner initialization;
- expose debug output.

Recommended during migration: log conflicts first, then move to strict once systems are stable.

## 24. Definition of done for v1

V1 is complete when:

- `esc_core` has tests for entity, component registry, archetype migration, query, command buffer, resources, runner validation;
- Velora can create match world through `esc_core`;
- all current gameplay systems run on component queries;
- snapshots are built from `esc_core` queries;
- stage-boundary command application matches current behavior;
- `runner.DebugAccess()` and `world.DebugArchetypes()` produce useful output;
- old full-world `QueryPlayerCells/QueryActiveNutrients` pattern is gone from active systems;
- `go test ./...` passes.
