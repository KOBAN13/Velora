package esc

type Position struct {
	X float32
	Y float32
}

type MoveDirection struct {
	X float32
	Y float32
}

type Owner struct {
	UserId uint64
}

type Health struct {
	HP int32
}

type Biomass struct {
	Value uint32
}

type Level struct {
	Value uint32
}

type Active struct {
	IsActive bool
}

type NutrientValue struct {
	Value uint32
}

type WallState struct {
	Open bool
}
