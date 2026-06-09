package systems

import (
	"Velora/esc"
	"Velora/server/Internal/server/match"
)

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
	var phaseSystem = NewPhaseSystem(runner.match)
	var inputSystem = NewInputSystem(runner.match)
	var movementSystem = NewMovementSystem(runner.match)
	var wallGateSystem = NewWallGateSystem(runner.match)
	var deathSystem = NewDeathSystem(runner.match)

	runner.initializeSystems = append(runner.initializeSystems, nutrientSystem)

	runner.updateSystems = append(runner.updateSystems, phaseSystem)
	runner.updateSystems = append(runner.updateSystems, inputSystem)
	runner.updateSystems = append(runner.updateSystems, movementSystem)
	runner.updateSystems = append(runner.updateSystems, nutrientSystem)
	runner.updateSystems = append(runner.updateSystems, wallGateSystem)
	runner.updateSystems = append(runner.updateSystems, deathSystem)
}

func (runner *SystemRunner) UpdateSystems(tick float64, world *esc.World) {
	for _, update := range runner.updateSystems {
		update.Update(tick, world)
	}
}

func (runner *SystemRunner) InitializeSystems(world *esc.World) error {
	runner.match.Mu.Lock()
	defer runner.match.Mu.Unlock()

	for _, initializeSystem := range runner.initializeSystems {
		var err = initializeSystem.Start(world)

		if err != nil {
			return err
		}
	}

	return nil
}
