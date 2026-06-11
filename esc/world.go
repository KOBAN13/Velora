package esc

import (
	"Velora/server/Internal/server/config"
)

type WorldCapacity struct {
	Entities int
}

type World struct {
	Entities map[EntityId]Entity
}

func NewWorld(capacity WorldCapacity) *World {
	return &World{
		Entities: make(map[EntityId]Entity, capacity.Entities),
	}
}

func (w *World) HasEntity(entity EntityId) bool {
	_, ok := w.Entities[entity]
	return ok
}

func (w *World) RemoveEntity(entity EntityId) {
	delete(w.Entities, entity)
}

func (w *World) PlayerCells() []*PlayerCell {
	result := make([]*PlayerCell, 0)
	for _, entity := range w.Entities {
		if typed, ok := entity.(*PlayerCell); ok {
			result = append(result, typed)
		}
	}
	return result
}

func (w *World) Cores() []*Core {
	result := make([]*Core, 0)
	for _, entity := range w.Entities {
		if typed, ok := entity.(*Core); ok {
			result = append(result, typed)
		}
	}
	return result
}

func (w *World) Nutrients() []*Nutrient {
	result := make([]*Nutrient, 0)
	for _, entity := range w.Entities {
		if typed, ok := entity.(*Nutrient); ok {
			result = append(result, typed)
		}
	}
	return result
}

func (w *World) Walls() []*Wall {
	result := make([]*Wall, 0)
	for _, entity := range w.Entities {
		if typed, ok := entity.(*Wall); ok {
			result = append(result, typed)
		}
	}
	return result
}

func (w *World) CreatePlayerCell(
	id EntityId,
	ownerId uint64,
	position Position,
	playerConfig config.PlayerCellConfig,
) {
	w.Entities[id] = &PlayerCell{
		Id:       id,
		OwnerId:  Owner{ownerId},
		Position: position,
		HP:       Health{int32(playerConfig.HP)},
		MaxHP:    Health{int32(playerConfig.MaxHP)},
		Biomass:  Biomass{uint32(playerConfig.Biomass)},
		Level:    Level{uint32(playerConfig.Level)},
		Active:   Active{playerConfig.Alive},
	}
}

func (w *World) CreateCore(
	id EntityId,
	ownerId uint64,
	position Position,
	coreConfig config.CoreConfig,
) {
	w.Entities[id] = &Core{
		Id:       id,
		OwnerId:  Owner{ownerId},
		Position: position,
		HP:       Health{int32(coreConfig.HP)},
		MaxHP:    Health{int32(coreConfig.MaxHP)},
	}
}

func (w *World) CreateNutrient(
	id EntityId,
	position Position,
	value uint32,
	active bool,
) {
	w.Entities[id] = &Nutrient{
		Id:       id,
		Position: position,
		Value:    NutrientValue{value},
		Active:   Active{active},
	}
}

func (w *World) CreateWall(id EntityId, wallConfig config.WallConfig) {
	w.Entities[id] = &Wall{
		Id:   id,
		Open: WallState{wallConfig.Open},
	}
}
