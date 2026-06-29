package systems

import (
	esc_core "github.com/KOBAN13/kukuruzka-esc/ecs"
)

type SystemRunner struct {
	stages            []esc_core.StageID
	runner            *esc_core.Runner
	initializeSystems []Initializer
}

func NewSystemRunner() *SystemRunner {
	stages := []esc_core.StageID{
		StagePhase,
		StageInput,
		StageMovement,
		StageSpawn,
		StageRules,
		StageCleanup,
	}

	return &SystemRunner{
		stages:            stages,
		runner:            esc_core.NewRunner(stages),
		initializeSystems: []Initializer{},
	}
}

func (runner *SystemRunner) BuildSystems(world *esc_core.World) error {
	runner.runner = esc_core.NewRunner(runner.stages)
	runner.initializeSystems = runner.initializeSystems[:0]

	phaseSystem, err := NewPhaseSystem()

	if err != nil {
		return err
	}

	inputSystem, err := NewInputSystem(world)

	if err != nil {
		return err
	}

	movementSystem, err := NewMovementSystem(world)

	if err != nil {
		return err
	}

	nutrientSystem, err := NewNutrientSystem(world)

	if err != nil {
		return err
	}

	wallGateSystem, err := NewWallGateSystem(world)

	if err != nil {
		return err
	}

	deathSystem, err := NewDeathSystem(world)

	if err != nil {
		return err
	}

	runner.runner.Add(phaseSystem)
	runner.runner.Add(inputSystem)
	runner.runner.Add(movementSystem)
	runner.runner.Add(nutrientSystem)
	runner.runner.Add(wallGateSystem)
	runner.runner.Add(deathSystem)

	runner.initializeSystems = append(runner.initializeSystems, nutrientSystem)

	return runner.runner.ValidateAccess()
}

func (runner *SystemRunner) UpdateSystems(ctx *esc_core.Context) error {
	var err = runner.runner.Update(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (runner *SystemRunner) InitializeSystems(ctx *esc_core.Context) error {
	for _, initializeSystem := range runner.initializeSystems {
		if err := initializeSystem.Start(ctx); err != nil {
			return err
		}
	}

	if err := ctx.Commands.Apply(ctx.World); err != nil {
		return err
	}

	ctx.Commands.Clear()

	return nil
}
