package config

import "Velora/server/Internal/objects"

type AppConfig struct {
	DatabaseUrl string
	UserTable   string

	GoogleSheets objects.SharedCollection[GoogleSheetsConfig]
	Game         GameConfig
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
	Value int
	Alive bool
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
	nutrientValue, err := raw.Int(KeyNutrientValue)
	if err != nil {
		return NutrientConfig{}, err
	}

	nutrientActive, err := raw.Bool(KeyNutrientAlive)
	if err != nil {
		return NutrientConfig{}, err
	}

	return NutrientConfig{
		Value: nutrientValue,
		Alive: nutrientActive,
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

func NewGameConfig(raw RawConfig) (GameConfig, error) {
	playerCell, err := NewPlayerCellConfig(raw)

	if err != nil {
		return GameConfig{}, err
	}

	core, err := NewCoreConfig(raw)

	if err != nil {
		return GameConfig{}, err
	}

	nutrient, err := NewNutrientConfig(raw)

	if err != nil {
		return GameConfig{}, err
	}

	wall, err := NewWallConfig(raw)

	if err != nil {
		return GameConfig{}, err
	}

	return GameConfig{
		PlayerCell: playerCell,
		Core:       core,
		Nutrient:   nutrient,
		Wall:       wall,
	}, nil
}
