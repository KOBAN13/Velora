package config

type ConfigKey string

const (
	KeyPlayerCellHP      ConfigKey = "player_cell.hp"
	KeyPlayerCellMaxHP   ConfigKey = "player_cell.max_hp"
	KeyPlayerCellBiomass ConfigKey = "player_cell.biomass"
	KeyPlayerCellLevel   ConfigKey = "player_cell.level"
	KeyPlayerCellAlive   ConfigKey = "player_cell.alive"

	KeyCoreHP    ConfigKey = "core.hp"
	KeyCoreMaxHP ConfigKey = "core.max_hp"

	KeyMaxNutrient        ConfigKey = "max.nutrients"
	KeySpawnInterval      ConfigKey = "spawn.interval"
	KeySpawnBatchNutrient ConfigKey = "spawn.batch"
	KeyNutrientAlive      ConfigKey = "nutrient.active"

	KeyWallOpen ConfigKey = "wall.open"
)

const (
	PlayerParametersTableId = iota
	CoreEntityTableId
	NutrientTableId
	WallsTableId
)
