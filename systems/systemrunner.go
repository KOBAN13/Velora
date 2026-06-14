package systems

import (
	"Velora/esc"
)

type SystemRunner struct {
	updateSystems     []System
	initializeSystems []Initializer
}

func NewSystemRunner() *SystemRunner {
	return &SystemRunner{
		updateSystems:     []System{},
		initializeSystems: []Initializer{},
	}
}

func (runner *SystemRunner) BuildSystems() {
	var nutrientSystem = NewNutrientSystem()
	var phaseSystem = NewPhaseSystem()
	var inputSystem = NewInputSystem()
	var movementSystem = NewMovementSystem()
	var wallGateSystem = NewWallGateSystem()
	var deathSystem = NewDeathSystem()

	runner.initializeSystems = append(runner.initializeSystems, nutrientSystem)

	runner.updateSystems = append(runner.updateSystems, phaseSystem)
	runner.updateSystems = append(runner.updateSystems, inputSystem)
	runner.updateSystems = append(runner.updateSystems, movementSystem)
	runner.updateSystems = append(runner.updateSystems, nutrientSystem)
	runner.updateSystems = append(runner.updateSystems, wallGateSystem)
	runner.updateSystems = append(runner.updateSystems, deathSystem)
}

func (runner *SystemRunner) UpdateSystems(ctx *esc.SystemContext, world *esc.World) {
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

			system.Update(ctx, world)
		}

		if ctx.Commands != nil {
			ctx.Commands.Execute(world, ctx.Resources)
			ctx.Commands.Clear()
		}
	}
}

func (runner *SystemRunner) InitializeSystems(ctx *esc.SystemContext, world *esc.World) error {
	for _, initializeSystem := range runner.initializeSystems {
		var err = initializeSystem.Start(ctx, world)

		if err != nil {
			return err
		}
	}

	if ctx.Commands != nil {
		ctx.Commands.Execute(world, ctx.Resources)
		ctx.Commands.Clear()
	}

	return nil
}
