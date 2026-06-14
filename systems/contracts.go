package systems

import (
	"Velora/esc"
)

type Stage int

const (
	StagePhase Stage = iota
	StageInput
	StageMovement
	StageSpawn
	StageRules
	StageCleanup
)

type System interface {
	Name() string
	Stage() Stage
	Update(ctx *esc.SystemContext, world *esc.World)
}

type Initializer interface {
	Start(ctx *esc.SystemContext, world *esc.World) error
}
