package esc

import (
	"Velora/server/Internal/server/config"

	esc_core "github.com/KOBAN13/kukuruzka-esc/ecs"
)

func NewPlayerCellBundle(
	ownerId uint64,
	position Position,
	playerConfig config.PlayerCellConfig,
) esc_core.Bundle {
	return esc_core.BundleFunc(func(b *esc_core.BundleBuilder) error {
		return b.
			With(PlayerTag{}).
			With(Owner{UserId: ownerId}).
			With(position).
			With(MoveDirection{}).
			With(Health{Value: int32(playerConfig.HP)}).
			With(MaxHealth{Value: int32(playerConfig.MaxHP)}).
			With(Biomass{Value: uint32(playerConfig.Biomass)}).
			With(Level{Value: uint32(playerConfig.Level)}).
			Err()
	})
}

func NewCoreBundle(
	ownerId uint64,
	position Position,
	coreConfig config.CoreConfig,
) esc_core.Bundle {
	return esc_core.BundleFunc(func(b *esc_core.BundleBuilder) error {
		return b.
			With(CoreTag{}).
			With(Owner{UserId: ownerId}).
			With(position).
			With(Health{Value: int32(coreConfig.HP)}).
			With(MaxHealth{Value: int32(coreConfig.MaxHP)}).
			Err()
	})
}

func NewNutrientBundle(
	position Position,
	value uint32,
	active bool,
) esc_core.Bundle {
	return esc_core.BundleFunc(func(b *esc_core.BundleBuilder) error {
		return b.
			With(NutrientTag{}).
			With(position).
			With(NutrientValue{Value: value}).
			With(Active{IsActive: active}).
			Err()
	})
}

func NewWallBundle(wallConfig config.WallConfig) esc_core.Bundle {
	return esc_core.BundleFunc(func(b *esc_core.BundleBuilder) error {
		return b.
			With(WallTag{}).
			With(WallState{Open: wallConfig.Open}).
			Err()
	})
}
