package match

import "time"

const (
	TickRate         = 20
	TickDuration     = 50 * time.Millisecond
	TimeDeltaSeconds = float32(0.05)
)

const (
	PrepareDuration = 3 * time.Second
	ActiveDuration  = 180 * time.Second
)

const (
	defaultNutrientMaxAttempts         = 1500
	defaultNutrientArenaHalfSize       = 22
	defaultNutrientMinPlayerDistance   = 4
	defaultNutrientMinCoreDistance     = 5
	defaultNutrientMinNutrientDistance = 2
)

const (
	BaseSpeed = 5
)
