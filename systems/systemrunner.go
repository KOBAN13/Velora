package systems

import (
	"Velora/esc"
	"Velora/server/Internal/server/match"
)

type SystemRunner struct {
	match *match.Match

	updateSystems     []System
	initializeSystems []Initializer
}

func NewSystemRunner(match *match.Match) *SystemRunner {
	return &SystemRunner{
		match:             match,
		updateSystems:     []System{},
		initializeSystems: []Initializer{},
	}
}

func (runner *SystemRunner) BuildSystems() {
	var nutrientSystem = NewNutrientSystem(runner.match)
	var phaseSystem = NewPhaseSystem(runner.match)
	var inputSystem = NewInputSystem(runner.match)
	var movementSystem = NewMovementSystem(runner.match)
	var wallGateSystem = NewWallGateSystem(runner.match)
	var deathSystem = NewDeathSystem()

	runner.initializeSystems = append(runner.initializeSystems, nutrientSystem)

	runner.updateSystems = append(runner.updateSystems, phaseSystem)
	runner.updateSystems = append(runner.updateSystems, inputSystem)
	runner.updateSystems = append(runner.updateSystems, movementSystem)
	runner.updateSystems = append(runner.updateSystems, nutrientSystem)
	runner.updateSystems = append(runner.updateSystems, wallGateSystem)
	runner.updateSystems = append(runner.updateSystems, deathSystem)
}

func (runner *SystemRunner) UpdateSystems(ctx *esc.SystemContext, world *esc.World) error {
	stages := []Stage{
		StagePhase,
		StageInput,
		StageMovement,
		StageSpawn,
		StageRules,
		StageCleanup,
	}

	for _, stage := range stages {
		for _, system := range runner.updateSystems {
			if system.Stage() != stage {
				continue
			}

			if err := system.Update(ctx, world); err != nil {
				return err
			}
		}

		if ctx.Commands != nil {
			ctx.Commands.Execute(world, ctx.Resources)
			ctx.Commands.Clear(world, ctx.Resources)
		}
	}

	return nil
}

func (runner *SystemRunner) InitializeSystems(ctx *esc.SystemContext, world *esc.World) error {
	runner.match.Mu.Lock()
	defer runner.match.Mu.Unlock()

	for _, initializeSystem := range runner.initializeSystems {
		var err = initializeSystem.Start(ctx, world)

		if err != nil {
			return err
		}
	}

	return nil
}
