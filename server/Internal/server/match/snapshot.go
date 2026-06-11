package match

import (
	"Velora/esc"
	"Velora/server/pkg/packets"
	"slices"
	"time"
)

func BuildMatchSnapshot(m *Match, world *esc.World, now time.Time) packets.Msg {
	var playerEntities = sortedEntitiesById(world.PlayerCells())
	var coreEntities = sortedEntitiesById(world.Cores())
	var nutrientEntities = sortedEntitiesById(world.Nutrients())
	var wallEntities = sortedEntitiesById(world.Walls())

	var playerCells = make([]*packets.PlayerCellEntityMessage, 0, len(playerEntities))
	var cores = make([]*packets.CoreEntityMessage, 0, len(coreEntities))
	var nutrients = make([]*packets.NutrientEntityMessage, 0, len(nutrientEntities))
	var walls = make([]*packets.WallEntityMessage, 0, len(wallEntities))

	for _, player := range playerEntities {
		playerCells = append(playerCells, &packets.PlayerCellEntityMessage{
			Id:       uint64(player.Id),
			OwnerId:  player.OwnerId.UserId,
			Position: newVector2Message(player.Position),
			Hp:       uint32(player.HP.HP),
			Biomass:  player.Biomass.Value,
			Level:    player.Level.Value,
			Alive:    player.Active.IsActive,
		})
	}

	for _, core := range coreEntities {
		cores = append(cores, &packets.CoreEntityMessage{
			Id:       uint64(core.Id),
			OwnerId:  core.OwnerId.UserId,
			Position: newVector2Message(core.Position),
			Hp:       uint32(core.HP.HP),
		})
	}

	for _, nutrient := range nutrientEntities {
		nutrients = append(nutrients, &packets.NutrientEntityMessage{
			Id:       uint64(nutrient.Id),
			Position: newVector2Message(nutrient.Position),
			Value:    nutrient.Value.Value,
			Active:   nutrient.Active.IsActive,
		})
	}

	for _, wall := range wallEntities {
		walls = append(walls, &packets.WallEntityMessage{
			Id:   uint64(wall.Id),
			Open: wall.Open.Open,
		})
	}

	return packets.NewMatchSnapshot(
		m.ID,
		m.ServerTick,
		m.Phase,
		phaseTimeLeftMs(m, now),
		playerCells,
		cores,
		nutrients,
		walls)
}

func sortedEntitiesById[T esc.Entity](entities []T) []T {
	slices.SortFunc(entities, func(a T, b T) int {
		if a.EntityID() < b.EntityID() {
			return -1
		}
		if a.EntityID() > b.EntityID() {
			return 1
		}
		return 0
	})

	return entities
}

func newVector2Message(position esc.Position) *packets.Vector2Message {
	return &packets.Vector2Message{
		X: position.X,
		Y: position.Y,
	}
}

func phaseTimeLeftMs(m *Match, now time.Time) int64 {
	if m.Phase == packets.MatchPhase_MATCH_PHASE_ENDED || m.PhaseEndsAt.IsZero() {
		return 0
	}

	var left = m.PhaseEndsAt.Sub(now)
	if left < 0 {
		return 0
	}

	return left.Milliseconds()
}
