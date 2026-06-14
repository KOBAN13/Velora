package esc

import (
	"math/rand/v2"
	"time"

	"Velora/server/pkg/packets"
)

type Resources struct {
	Inputs          *InputResource
	NutrientSpawner *NutrientSpawnerResource
	EntityIds       EntityAllocator
}

type InputResource struct {
	Inputs map[uint64]PlayerInput
}

type PlayerInput struct {
	MoveX      float32
	MoveY      float32
	ReceivedAt time.Time
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

type SystemContext struct {
	Tick         uint64
	DeltaSeconds float32
	Now          time.Time

	Phase       packets.MatchPhase
	PhaseEndsAt time.Time

	Commands  *CommandBuffer
	Resources *Resources
}
