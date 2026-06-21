package esc_core

import "fmt"

type Runner struct {
	stages  []StageID
	systems []System
}

func NewRunner(stages []StageID) *Runner {
	return &Runner{
		stages:  stages,
		systems: make([]System, 0, len(stages)),
	}
}

func (r *Runner) Add(system System) {
	r.systems = append(r.systems, system)
}

func (r *Runner) ValidateAccess() error {
	var accessByStage = make(map[StageID]AccessSet, len(r.stages))
	var ownerByStage = make(map[StageID]map[ComponentID]string, len(r.stages))

	for _, stage := range r.stages {
		accessByStage[stage] = NewAccessSet()
		ownerByStage[stage] = make(map[ComponentID]string)
	}

	for _, system := range r.systems {
		var stage = system.Stage()

		var stageAccess, ok = accessByStage[stage]

		if !ok {
			return fmt.Errorf("unknown stage %d for system %s", stage, system.Name())
		}

		var conflicts = stageAccess.ConflictsWith(system.Access())

		if len(conflicts) > 0 {
			var components = conflicts[0]

			return AccessConflict{
				Stage:     stage,
				Component: components,
				First:     ownerByStage[stage][components],
				Second:    system.Name(),
			}
		}

		stageAccess.Merge(system.Access())

		for component := range system.Access().Reads {
			if _, exists := ownerByStage[stage][component]; !exists {
				ownerByStage[stage][component] = system.Name()
			}
		}

		for component := range system.Access().Writes {
			ownerByStage[stage][component] = system.Name()
		}
	}

	return nil
}

func (r *Runner) Update(ctx *Context) error {
	for _, stage := range r.stages {
		ctx.Stage = stage

		ctx.World.mutationPhase = MutationRunningSystem

		for _, system := range r.systems {
			if system.Stage() != stage {
				continue
			}

			var err = system.Update(ctx)

			if err != nil {
				return err
			}
		}

		ctx.World.mutationPhase = MutationApplyingCommands

		//Применяем command buffer
		//Чистим command buffer

		ctx.World.mutationPhase = MutationIdle
	}

	return nil
}
