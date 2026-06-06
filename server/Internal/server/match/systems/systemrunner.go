package systems

import "Velora/server/Internal/server/match"

type SystemRunner struct {
	match *match.Match

	updateSystems     []Update
	initializeSystems []Initialize
}

func NewSystemRunner(match *match.Match) *SystemRunner {
	return &SystemRunner{
		match:             match,
		updateSystems:     []Update{},
		initializeSystems: []Initialize{},
	}
}

func (runner *SystemRunner) BuildSystems() {
	var nutrientSystem = NewNutrientSystem(runner.match)

	runner.initializeSystems = append(runner.initializeSystems, nutrientSystem)

	runner.updateSystems = append(runner.updateSystems, nutrientSystem)
}

func (runner *SystemRunner) UpdateSystems(tick float64, world *match.World) {
	for _, update := range runner.updateSystems {
		update.Update(tick, world)
	}
}

func (runner *SystemRunner) InitializeSystems(world *match.World) error {
	for _, initializeSystem := range runner.initializeSystems {
		var err = initializeSystem.Start(world)

		if err != nil {
			return err
		}
	}

	return nil
}
