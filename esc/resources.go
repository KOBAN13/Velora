package esc

import (
	"Velora/server/pkg/packets"
	"math/rand/v2"
	"time"
)

type PlayerInputSlice []PlayerInput

type PlayerInput struct {
	MoveX      float32
	MoveY      float32
	ReceivedAt time.Time
}

type MatchPhaseResource struct {
	Phase       packets.MatchPhase
	PhaseEndsAt time.Time
	Now         time.Time
}

type NutrientSpawnerResource struct {
	Rng *rand.Rand

	LastSpawnTick float64
	SpawnInterval float64
	MaxNutrients  int
	SpawnBatch    int
	MaxAttempts   int
	NutrientValue uint32

	ArenaHalfSize       float32
	MinPlayerDistance   float32
	MinCoreDistance     float32
	MinNutrientDistance float32

	NutrientActive bool
}
