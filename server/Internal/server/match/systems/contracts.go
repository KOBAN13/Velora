package systems

import "Velora/server/Internal/server/match"

type Update interface {
	Update(tick float64, world *match.World)
}

type Initialize interface {
	Start(world *match.World) error
}
