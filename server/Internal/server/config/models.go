package config

import "Velora/server/Internal/objects"

type AppConfig struct {
	DatabaseUrl string
	UserTable   string

	GoogleSheets *objects.SharedCollection[GoogleSheetsConfig]
	Game         *GameConfig
}

type GoogleSheetsConfig struct {
	SpreadsheetId      string
	ConfigSheetName    string
	ServiceAccountFile string
}

type GameConfig struct {
	PlayerCell PlayerCellConfig
	Core       CoreConfig
	Nutrient   NutrientConfig
	Wall       WallConfig
}

type PlayerCellConfig struct {
	HP      int
	MaxHP   int
	Biomass int
	Level   int
	Alive   bool
}

type CoreConfig struct {
	HP    int
	MaxHP int
}

type NutrientConfig struct {
	MaxNutrients  int
	SpawnInterval float64
	SpawnBatch    int
	Alive         bool
}

type WallConfig struct {
	Open bool
}

func NewPlayerCellConfig(raw RawConfig) (PlayerCellConfig, error) {
	hp, err := raw.Int(KeyPlayerCellHP)
	if err != nil {
		return PlayerCellConfig{}, err
	}

	maxHP, err := raw.Int(KeyPlayerCellMaxHP)
	if err != nil {
		return PlayerCellConfig{}, err
	}

	biomass, err := raw.Int(KeyPlayerCellBiomass)
	if err != nil {
		return PlayerCellConfig{}, err
	}

	level, err := raw.Int(KeyPlayerCellLevel)
	if err != nil {
		return PlayerCellConfig{}, err
	}

	alive, err := raw.Bool(KeyPlayerCellAlive)
	if err != nil {
		return PlayerCellConfig{}, err
	}

	return PlayerCellConfig{
		HP:      hp,
		MaxHP:   maxHP,
		Biomass: biomass,
		Level:   level,
		Alive:   alive,
	}, nil
}

func NewCoreConfig(raw RawConfig) (CoreConfig, error) {
	hp, err := raw.Int(KeyCoreHP)
	if err != nil {
		return CoreConfig{}, err
	}

	maxHP, err := raw.Int(KeyCoreMaxHP)
	if err != nil {
		return CoreConfig{}, err
	}

	return CoreConfig{
		HP:    hp,
		MaxHP: maxHP,
	}, nil
}

func NewNutrientConfig(raw RawConfig) (NutrientConfig, error) {
	nutrientMax, err := raw.Int(KeyMaxNutrient)
	if err != nil {
		return NutrientConfig{}, err
	}

	nutrientActive, err := raw.Bool(KeyNutrientAlive)
	if err != nil {
		return NutrientConfig{}, err
	}

	nutrientSpawnInterval, err := raw.Float(KeySpawnInterval)
	if err != nil {
		return NutrientConfig{}, err
	}

	nutrientSpawnBatch, err := raw.Int(KeySpawnBatchNutrient)
	if err != nil {
		return NutrientConfig{}, err
	}

	return NutrientConfig{
		MaxNutrients:  nutrientMax,
		SpawnInterval: nutrientSpawnInterval,
		SpawnBatch:    nutrientSpawnBatch,
		Alive:         nutrientActive,
	}, nil
}

func NewWallConfig(raw RawConfig) (WallConfig, error) {
	wallOpen, err := raw.Bool(KeyWallOpen)
	if err != nil {
		return WallConfig{}, err
	}

	return WallConfig{
		Open: wallOpen,
	}, err
}
