package match

import (
	"Velora/esc"
	"Velora/server/pkg/packets"
	"cmp"
	"slices"
	"time"

	esc_core "github.com/KOBAN13/kukuruzka-esc/ecs"
)

type SnapshotQueries struct {
	players   *esc_core.Query
	cores     *esc_core.Query
	nutrients *esc_core.Query
	walls     *esc_core.Query
}

func BuildMatchSnapshot(m *Match, queries *SnapshotQueries, now time.Time) packets.Msg {
	playerCells, err := buildPlayerCellMessage(queries.players)
	if err != nil {
		return nil
	}

	cores, err := buildCoreMessage(queries.cores)
	if err != nil {
		return nil
	}

	nutrients, err := buildNutrientMessage(queries.nutrients)
	if err != nil {
		return nil
	}

	walls, err := buildWallsMessage(queries.walls)
	if err != nil {
		return nil
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

func buildWallsMessage(walls *esc_core.Query) ([]*packets.WallEntityMessage, error) {
	var it = walls.Iter()

	var wallsMessage = make([]*packets.WallEntityMessage, 0)

	for it.Next() {
		var entity = it.Entity()

		open, err := esc_core.Read[esc.WallState](it)
		if err != nil {
			return nil, err
		}

		wallsMessage = append(wallsMessage, &packets.WallEntityMessage{
			Id:   uint64(entity.Index()),
			Open: open.Open,
		})
	}

	slices.SortFunc(wallsMessage, func(a, b *packets.WallEntityMessage) int {
		return cmp.Compare(a.Id, b.Id)
	})

	return wallsMessage, nil
}

func buildNutrientMessage(nutrients *esc_core.Query) ([]*packets.NutrientEntityMessage, error) {
	var it = nutrients.Iter()

	var nutrientsMessage = make([]*packets.NutrientEntityMessage, 0)

	for it.Next() {
		var entity = it.Entity()

		position, err := esc_core.Read[esc.Position](it)
		if err != nil {
			return nil, err
		}

		value, err := esc_core.Read[esc.NutrientValue](it)
		if err != nil {
			return nil, err
		}

		active, err := esc_core.Read[esc.Active](it)
		if err != nil {
			return nil, err
		}

		nutrientsMessage = append(nutrientsMessage, &packets.NutrientEntityMessage{
			Id:       uint64(entity.Index()),
			Position: newVector2Message(position),
			Value:    value.Value,
			Active:   active.IsActive,
		})
	}

	slices.SortFunc(nutrientsMessage, func(a, b *packets.NutrientEntityMessage) int {
		return cmp.Compare(a.Id, b.Id)
	})

	return nutrientsMessage, nil
}

func buildCoreMessage(cores *esc_core.Query) ([]*packets.CoreEntityMessage, error) {
	var it = cores.Iter()

	var coresMessage = make([]*packets.CoreEntityMessage, 0)

	for it.Next() {
		var entity = it.Entity()

		owner, err := esc_core.Read[esc.Owner](it)
		if err != nil {
			return nil, err
		}

		position, err := esc_core.Read[esc.Position](it)
		if err != nil {
			return nil, err
		}

		health, err := esc_core.Read[esc.Health](it)
		if err != nil {
			return nil, err
		}

		coresMessage = append(coresMessage, &packets.CoreEntityMessage{
			Id:       uint64(entity.Index()),
			OwnerId:  owner.UserId,
			Position: newVector2Message(position),
			Hp:       uint32(health.Value),
		})
	}

	slices.SortFunc(coresMessage, func(a, b *packets.CoreEntityMessage) int {
		return cmp.Compare(a.Id, b.Id)
	})

	return coresMessage, nil
}

func buildPlayerCellMessage(players *esc_core.Query) ([]*packets.PlayerCellEntityMessage, error) {
	var it = players.Iter()

	var playerCells = make([]*packets.PlayerCellEntityMessage, 0)

	for it.Next() {
		var entity = it.Entity()

		owner, err := esc_core.Read[esc.Owner](it)
		if err != nil {
			return nil, err
		}

		position, err := esc_core.Read[esc.Position](it)
		if err != nil {
			return nil, err
		}

		health, err := esc_core.Read[esc.Health](it)
		if err != nil {
			return nil, err
		}

		biomass, err := esc_core.Read[esc.Biomass](it)
		if err != nil {
			return nil, err
		}

		level, err := esc_core.Read[esc.Level](it)
		if err != nil {
			return nil, err
		}

		active, err := esc_core.Read[esc.Active](it)
		if err != nil {
			return nil, err
		}

		playerCells = append(playerCells, &packets.PlayerCellEntityMessage{
			Id:       uint64(entity.Index()),
			OwnerId:  owner.UserId,
			Position: newVector2Message(position),
			Hp:       uint32(health.Value),
			Biomass:  biomass.Value,
			Level:    level.Value,
			Alive:    active.IsActive,
		})
	}

	slices.SortFunc(playerCells, func(a, b *packets.PlayerCellEntityMessage) int {
		return cmp.Compare(a.Id, b.Id)
	})

	return playerCells, nil
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
