package systems

import (
	esc_core "github.com/KOBAN13/kukuruzka-esc/ecs"
)

const (
	StagePhase esc_core.StageID = iota
	StageInput
	StageMovement
	StageSpawn
	StageRules
	StageCleanup
)

type Initializer interface {
	Start(ctx *esc_core.Context, world *esc_core.World) error
}
