package systems

import (
	"errors"

	"Velora/esc"

	esc_core "github.com/KOBAN13/kukuruzka-esc/ecs"
)

var (
	ErrNutrientSpawn = errors.New("nutrient spawn failed")
)

const (
	DefaultNutrientPickUpDistance = 1.5
)

type NutrientSystem struct {
	players   *esc_core.Query
	nutrients *esc_core.Query
	cores     *esc_core.Query
}

func NewNutrientSystem(world *esc_core.World) (*NutrientSystem, error) {
	players, err := esc_core.
		NewQuery(world, "NutrientPlayers").
		With(esc_core.Component[esc.PlayerTag]()).
		Read(esc_core.Component[esc.Active]()).
		Read(esc_core.Component[esc.Position]()).
		Write(esc_core.Component[esc.Biomass]()).
		Write(esc_core.Component[esc.Level]()).
		Build()

	if err != nil {
		return nil, err
	}

	nutrients, err := esc_core.
		NewQuery(world, "Nutrients").
		With(esc_core.Component[esc.NutrientTag]()).
		Write(esc_core.Component[esc.Position]()).
		Write(esc_core.Component[esc.NutrientValue]()).
		Write(esc_core.Component[esc.Active]()).
		Build()

	if err != nil {
		return nil, err
	}

	cores, err := esc_core.
		NewQuery(world, "NutrientCores").
		With(esc_core.Component[esc.CoreTag]()).
		Read(esc_core.Component[esc.Position]()).
		Build()

	if err != nil {
		return nil, err
	}

	return &NutrientSystem{
		players:   players,
		nutrients: nutrients,
		cores:     cores,
	}, nil
}

func (*NutrientSystem) Name() string {
	return "NutrientSystem"
}

func (*NutrientSystem) Stage() esc_core.StageID {
	return StageSpawn
}

func (s *NutrientSystem) Access() esc_core.AccessSet {
	var access = s.players.Access()
	access.Merge(s.nutrients.Access())
	access.Merge(s.cores.Access())
	return access
}

func (s *NutrientSystem) DebugQueries() []esc_core.QueryDebugInfo {
	return []esc_core.QueryDebugInfo{
		s.players.DebugInfo(),
		s.nutrients.DebugInfo(),
		s.cores.DebugInfo(),
	}
}

func (s *NutrientSystem) Start(ctx *esc_core.Context) error {
	spawner, err := esc_core.GetResources[esc.NutrientSpawnerResource](ctx.Resources)

	if err != nil {
		return err
	}

	activeCount, totalCount, inactive, err := s.nutrientStats()

	if err != nil {
		return err
	}

	missing := spawner.MaxNutrients - activeCount

	if missing <= 0 {
		return nil
	}

	spawned, err := s.spawn(ctx, spawner, missing, inactive, spawner.MaxNutrients-totalCount)

	if err != nil {
		return err
	}

	if spawned != missing {
		return ErrNutrientSpawn
	}

	return nil
}

func (s *NutrientSystem) Update(ctx *esc_core.Context) error {
	if err := s.pickUpNutrients(); err != nil {
		return err
	}

	spawner, err := esc_core.GetResources[esc.NutrientSpawnerResource](ctx.Resources)

	if err != nil {
		return err
	}

	tick := float64(ctx.Tick)

	if spawner.SpawnInterval > 0 && tick-spawner.LastSpawnTick < spawner.SpawnInterval {
		return nil
	}

	spawner.LastSpawnTick = tick

	activeCount, totalCount, inactive, err := s.nutrientStats()

	if err != nil {
		return err
	}

	missing := spawner.MaxNutrients - activeCount

	if missing <= 0 || spawner.SpawnBatch <= 0 {
		return nil
	}

	_, err = s.spawn(ctx, spawner, min(missing, spawner.SpawnBatch), inactive, spawner.MaxNutrients-totalCount)
	return err
}

func (s *NutrientSystem) pickUpNutrients() error {
	players := s.players.Iter()

	for players.Next() {
		active, err := esc_core.Read[esc.Active](players)

		if err != nil {
			return err
		}

		if !active.IsActive {
			continue
		}

		position, err := esc_core.Read[esc.Position](players)

		if err != nil {
			return err
		}

		biomass, err := esc_core.Write[esc.Biomass](players)

		if err != nil {
			return err
		}

		level, err := esc_core.Write[esc.Level](players)

		if err != nil {
			return err
		}

		if err := s.pickUpNutrientsForPlayer(position, biomass, level); err != nil {
			return err
		}
	}

	return nil
}

func (s *NutrientSystem) pickUpNutrientsForPlayer(
	playerPosition esc.Position,
	playerBiomass *esc.Biomass,
	playerLevel *esc.Level,
) error {
	nutrients := s.nutrients.Iter()

	for nutrients.Next() {
		active, err := esc_core.Write[esc.Active](nutrients)

		if err != nil {
			return err
		}

		if !active.IsActive {
			continue
		}

		position, err := esc_core.Write[esc.Position](nutrients)

		if err != nil {
			return err
		}

		if distanceSquared(playerPosition, *position) > square(DefaultNutrientPickUpDistance) {
			continue
		}

		value, err := esc_core.Write[esc.NutrientValue](nutrients)

		if err != nil {
			return err
		}

		active.IsActive = false
		playerBiomass.Value += value.Value
		playerLevel.Value = 1 + playerBiomass.Value/100
	}

	return nil
}
