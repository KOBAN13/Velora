package systems

import (
	esc_core "github.com/KOBAN13/kukuruzka-esc/ecs"
)

const (
	StagePhase esc_core.StageID = iota + 1
	StageInput
	StageMovement
	StageSpawn
	StageRules
	StageCleanup
)

type Initializer interface {
	Start(ctx *esc_core.Context) error
}
