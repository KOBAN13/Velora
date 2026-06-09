package esc

import (
	"math"
)

func (d MoveDirection) IsZero() bool {
	return d.X == 0 && d.Y == 0
}

func Zero() MoveDirection {
	return MoveDirection{
		X: 0,
		Y: 0,
	}
}

func (d MoveDirection) Length() float32 {
	return float32(math.Sqrt(float64(d.X*d.X + d.Y*d.Y)))
}

func (d MoveDirection) Normalize() MoveDirection {
	length := d.Length()
	if length == 0 {
		return MoveDirection{}
	}

	return MoveDirection{
		X: d.X / length,
		Y: d.Y / length,
	}
}
