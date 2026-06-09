package esc

import (
	"math"
)

func (a Position) DistanceTo(b Position) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y

	return math.Sqrt(float64(dx*dx + dy*dy))
}
