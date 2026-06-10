package esc

import (
	"Velora/server/Internal/server/config"
)

type WorldCapacity struct {
	PlayerCells int
	Cores       int
	Nutrients   int
	Walls       int
}

type World struct {
	PlayerCells map[EntityId]PlayerCell
	Cores       map[EntityId]Core
	Nutrients   map[EntityId]Nutrient
	Walls       map[EntityId]Wall

	Directions     map[EntityId]MoveDirection
	Positions      map[EntityId]Position
	Owners         map[EntityId]Owner
	Health         map[EntityId]Health
	MaxHealth      map[EntityId]Health
	Biomass        map[EntityId]Biomass
	Levels         map[EntityId]Level
	Active         map[EntityId]Active
	NutrientValues map[EntityId]NutrientValue
	WallStates     map[EntityId]WallState
}

func NewWorld(capacity WorldCapacity) *World {
	return &World{
		PlayerCells: make(map[EntityId]PlayerCell, capacity.PlayerCells),
		Cores:       make(map[EntityId]Core, capacity.Cores),
		Nutrients:   make(map[EntityId]Nutrient, capacity.Nutrients),
		Walls:       make(map[EntityId]Wall, capacity.Walls),

		Directions:     make(map[EntityId]MoveDirection, capacity.PlayerCells),
		Positions:      make(map[EntityId]Position, capacity.PlayerCells+capacity.Cores+capacity.Nutrients),
		Owners:         make(map[EntityId]Owner, capacity.PlayerCells+capacity.Cores),
		Health:         make(map[EntityId]Health, capacity.PlayerCells+capacity.Cores),
		MaxHealth:      make(map[EntityId]Health, capacity.PlayerCells+capacity.Cores),
		Biomass:        make(map[EntityId]Biomass, capacity.PlayerCells),
		Levels:         make(map[EntityId]Level, capacity.PlayerCells),
		Active:         make(map[EntityId]Active, capacity.PlayerCells+capacity.Nutrients),
		NutrientValues: make(map[EntityId]NutrientValue, capacity.Nutrients),
		WallStates:     make(map[EntityId]WallState, capacity.Walls),
	}
}

func (w *World) HasEntity(entity EntityId) bool {
	if _, ok := w.PlayerCells[entity]; ok {
		return true
	}

	if _, ok := w.Cores[entity]; ok {
		return true
	}

	if _, ok := w.Nutrients[entity]; ok {
		return true
	}

	if _, ok := w.Walls[entity]; ok {
		return true
	}

	if _, ok := w.Directions[entity]; ok {
		return true
	}

	if _, ok := w.Positions[entity]; ok {
		return true
	}

	if _, ok := w.Owners[entity]; ok {
		return true
	}

	if _, ok := w.Health[entity]; ok {
		return true
	}

	if _, ok := w.MaxHealth[entity]; ok {
		return true
	}

	if _, ok := w.Biomass[entity]; ok {
		return true
	}

	if _, ok := w.Levels[entity]; ok {
		return true
	}

	if _, ok := w.Active[entity]; ok {
		return true
	}

	if _, ok := w.NutrientValues[entity]; ok {
		return true
	}

	if _, ok := w.WallStates[entity]; ok {
		return true
	}

	return false
}

func (w *World) RemoveEntity(entity EntityId) {
	delete(w.PlayerCells, entity)
	delete(w.Cores, entity)
	delete(w.Nutrients, entity)
	delete(w.Walls, entity)
	delete(w.Directions, entity)
	delete(w.Positions, entity)
	delete(w.Owners, entity)
	delete(w.Health, entity)
	delete(w.MaxHealth, entity)
	delete(w.Biomass, entity)
	delete(w.Levels, entity)
	delete(w.Active, entity)
	delete(w.NutrientValues, entity)
	delete(w.WallStates, entity)
}

func (w *World) CreatePlayerCell(
	id EntityId,
	ownerId uint64,
	position Position,
	playerConfig config.PlayerCellConfig,
) {
	w.PlayerCells[id] = PlayerCell{}
	w.Positions[id] = position
	w.Owners[id] = Owner{UserId: ownerId}
	w.Health[id] = Health{HP: int32(playerConfig.HP)}
	w.MaxHealth[id] = Health{HP: int32(playerConfig.MaxHP)}
	w.Biomass[id] = Biomass{Value: uint32(playerConfig.Biomass)}
	w.Levels[id] = Level{Value: uint32(playerConfig.Level)}
	w.Active[id] = Active{IsActive: playerConfig.Alive}
}

func (w *World) CreateCore(
	id EntityId,
	ownerId uint64,
	position Position,
	coreConfig config.CoreConfig,
) {
	w.Cores[id] = Core{}
	w.Positions[id] = position
	w.Owners[id] = Owner{UserId: ownerId}
	w.Health[id] = Health{HP: int32(coreConfig.HP)}
	w.MaxHealth[id] = Health{HP: int32(coreConfig.MaxHP)}
}

func (w *World) CreateNutrient(
	id EntityId,
	position Position,
	value uint32,
	active bool,
) {
	w.Nutrients[id] = Nutrient{}
	w.Positions[id] = position
	w.NutrientValues[id] = NutrientValue{Value: value}
	w.Active[id] = Active{IsActive: active}
}

func (w *World) CreateWall(id EntityId, wallConfig config.WallConfig) {
	w.Walls[id] = Wall{}
	w.WallStates[id] = WallState{Open: wallConfig.Open}
}
