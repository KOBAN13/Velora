package match

import (
	"Velora/server/pkg/packets"
	"slices"
	"time"
)

func BuildMatchSnapshot(m *Match, world *World, now time.Time) packets.Msg {
	var playerCells = make([]*packets.PlayerCellEntityMessage, 0, len(world.PlayerCells))
	var cores = make([]*packets.CoreEntityMessage, 0, len(world.Cores))
	var nutrients = make([]*packets.NutrientEntityMessage, 0, len(world.Nutrients))
	var walls = make([]*packets.WallEntityMessage, 0, len(world.Walls))

	for _, id := range sortedEntityIds(world.PlayerCells) {
		var position = world.Positions[id]
		var owner = world.Owners[id]
		var health = world.Health[id]
		var biomass = world.Biomass[id]
		var level = world.Levels[id]
		var active = world.Active[id]

		playerCells = append(playerCells, &packets.PlayerCellEntityMessage{
			Id:       uint64(id),
			OwnerId:  owner.UserId,
			Position: newVector2Message(position),
			Hp:       uint32(health.HP),
			Biomass:  biomass.Value,
			Level:    level.Value,
			Alive:    active.IsActive,
		})
	}

	for _, id := range sortedEntityIds(world.Cores) {
		var position = world.Positions[id]
		var owner = world.Owners[id]
		var health = world.Health[id]

		cores = append(cores, &packets.CoreEntityMessage{
			Id:       uint64(id),
			OwnerId:  owner.UserId,
			Position: newVector2Message(position),
			Hp:       uint32(health.HP),
		})
	}

	for _, id := range sortedEntityIds(world.Nutrients) {
		var position = world.Positions[id]
		var value = world.NutrientValues[id]
		var active = world.Active[id]

		nutrients = append(nutrients, &packets.NutrientEntityMessage{
			Id:       uint64(id),
			Position: newVector2Message(position),
			Value:    value.Value,
			Active:   active.IsActive,
		})
	}

	for _, id := range sortedEntityIds(world.Walls) {
		var state = world.WallStates[id]

		walls = append(walls, &packets.WallEntityMessage{
			Id:   uint64(id),
			Open: state.Open,
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

func sortedEntityIds[T any](entities map[EntityId]T) []EntityId {
	var ids = make([]EntityId, 0, len(entities))

	for id := range entities {
		ids = append(ids, id)
	}

	slices.Sort(ids)

	return ids
}

func newVector2Message(position Position) *packets.Vector2Message {
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
