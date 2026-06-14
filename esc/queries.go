package esc

func (w *World) QueryPlayerCells() []*PlayerCell {
	result := make([]*PlayerCell, 0)
	for _, entity := range w.Entities {
		if typed, ok := entity.(*PlayerCell); ok {
			result = append(result, typed)
		}
	}
	return result
}

func (w *World) QueryActivePlayerCells() []*PlayerCell {
	result := make([]*PlayerCell, 0)
	for _, entity := range w.Entities {
		if typed, ok := entity.(*PlayerCell); ok {
			if typed.Active.IsActive != false {
				result = append(result, typed)
			}
		}
	}
	return result
}

func (w *World) QueryCores() []*Core {
	result := make([]*Core, 0)
	for _, entity := range w.Entities {
		if typed, ok := entity.(*Core); ok {
			result = append(result, typed)
		}
	}
	return result
}

func (w *World) QueryNutrients() []*Nutrient {
	result := make([]*Nutrient, 0)
	for _, entity := range w.Entities {
		if typed, ok := entity.(*Nutrient); ok {
			result = append(result, typed)
		}
	}
	return result
}

func (w *World) QueryActiveNutrients() []*Nutrient {
	result := make([]*Nutrient, 0)
	for _, entity := range w.Entities {
		if typed, ok := entity.(*Nutrient); ok {
			if typed.Active.IsActive != false {
				result = append(result, typed)
			}
		}
	}
	return result
}

func (w *World) QueryInactiveNutrients() []*Nutrient {
	result := make([]*Nutrient, 0)
	for _, entity := range w.Entities {
		if typed, ok := entity.(*Nutrient); ok {
			if typed.Active.IsActive == false {
				result = append(result, typed)
			}
		}
	}
	return result
}

func (w *World) QueryWalls() []*Wall {
	result := make([]*Wall, 0)
	for _, entity := range w.Entities {
		if typed, ok := entity.(*Wall); ok {
			result = append(result, typed)
		}
	}
	return result
}

func (w *World) SetPosition(id EntityId, value Position) {
	entity, ok := w.Entities[id]
	if !ok {
		return
	}

	switch typed := entity.(type) {
	case *PlayerCell:
		typed.Position = value
	case *Core:
		typed.Position = value
	case *Nutrient:
		typed.Position = value
	}
}

func (w *World) SetDirection(id EntityId, value MoveDirection) {
	entity, ok := w.Entities[id]
	if !ok {
		return
	}

	if typed, ok := entity.(*PlayerCell); ok {
		typed.Direction = value
	}
}

func (w *World) SetActive(id EntityId, value Active) {
	entity, ok := w.Entities[id]
	if !ok {
		return
	}

	switch typed := entity.(type) {
	case *PlayerCell:
		typed.Active = value
	case *Nutrient:
		typed.Active = value
	}
}

func (w *World) SetHealth(id EntityId, value Health) {
	entity, ok := w.Entities[id]
	if !ok {
		return
	}

	switch typed := entity.(type) {
	case *PlayerCell:
		typed.HP = value
	case *Core:
		typed.HP = value
	}
}

func (w *World) SetBiomass(id EntityId, value Biomass) {
	entity, ok := w.Entities[id]
	if !ok {
		return
	}

	if typed, ok := entity.(*PlayerCell); ok {
		typed.Biomass = value
	}
}

func (w *World) SetLevel(id EntityId, value Level) {
	entity, ok := w.Entities[id]
	if !ok {
		return
	}

	if typed, ok := entity.(*PlayerCell); ok {
		typed.Level = value
	}
}

func (w *World) SetNutrient(id EntityId, position Position, value uint32, active bool) {
	entity, ok := w.Entities[id]
	if !ok {
		return
	}

	if typed, ok := entity.(*Nutrient); ok {
		typed.Position = position
		typed.Value = NutrientValue{value}
		typed.Active = Active{active}
	}
}

func (w *World) SetWallState(id EntityId, value WallState) {
	entity, ok := w.Entities[id]
	if !ok {
		return
	}

	if typed, ok := entity.(*Wall); ok {
		typed.Open = value
	}
}
