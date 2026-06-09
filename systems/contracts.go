package systems

import (
	"Velora/esc"
)

type Update interface {
	Update(tick float64, world *esc.World)
}

type Initialize interface {
	Start(world *esc.World) error
}
