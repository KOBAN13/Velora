package config

import (
	"Velora/server/Internal/objects"
	"context"
	"fmt"

	"google.golang.org/api/sheets/v4"
)

type GetEnvFunc func(string) string

type gameConfigSheetSpec struct {
	envKey string
	apply  func(RawConfig, *GameConfig) error
}

var gameConfigSheetSpecs = []gameConfigSheetSpec{
	{
		envKey: "GOOGLE_SHEETS_CONFIG_SHEET_PLAYER_PARAMETERS",
		apply:  applyPlayerCellConfig,
	},
	{
		envKey: "GOOGLE_SHEETS_CONFIG_SHEET_CORE_ENTITY",
		apply:  applyCoreConfig,
	},
	{
		envKey: "GOOGLE_SHEETS_CONFIG_SHEET_NUTRIENT",
		apply:  applyNutrientConfig,
	},
	{
		envKey: "GOOGLE_SHEETS_CONFIG_SHEET_PLAYER_WALLS",
		apply:  applyWallConfig,
	},
}

func NewAppConfig(ctx context.Context, sheetService *sheets.Service, getEnv GetEnvFunc) (*AppConfig, error) {
	gameConfig, err := loadGameConfig(
		ctx,
		sheetService,
		getEnv("GOOGLE_SHEETS_SPREADSHEET_ID"),
		getEnv)

	if err != nil {
		return nil, err
	}

	return &AppConfig{
		DatabaseUrl:  getEnv("DATABASE_URL"),
		UserTable:    getEnv("USERS_TABLE"),
		GoogleSheets: buildGoogleSheetsConfigs(getEnv),
		Game:         gameConfig,
	}, nil
}

func buildGoogleSheetsConfigs(getEnv GetEnvFunc) *objects.SharedCollection[GoogleSheetsConfig] {
	var result = objects.NewSharedCollection[GoogleSheetsConfig]()

	for _, spec := range gameConfigSheetSpecs {
		result.Add(GoogleSheetsConfig{
			SpreadsheetId:      getEnv("GOOGLE_SHEETS_SPREADSHEET_ID"),
			ConfigSheetName:    getEnv(spec.envKey),
			ServiceAccountFile: getEnv("GOOGLE_SERVICE_ACCOUNT_FILE"),
		})
	}

	return result
}

func loadGameConfig(
	ctx context.Context,
	sheetService *sheets.Service,
	spreadsheetID string,
	getEnv GetEnvFunc) (*GameConfig, error) {
	var gameConfig = &GameConfig{}

	for _, spec := range gameConfigSheetSpecs {
		var sheetName = getEnv(spec.envKey)

		rawConfig, err := LoadKeyValueConfig(ctx, sheetService, spreadsheetID, sheetName)

		if err != nil {
			return nil, fmt.Errorf("load sheet %s: %w", sheetName, err)
		}

		if err := spec.apply(rawConfig, gameConfig); err != nil {
			return nil, fmt.Errorf("parse sheet %s: %w", sheetName, err)
		}
	}

	return gameConfig, nil
}

func applyPlayerCellConfig(raw RawConfig, gameConfig *GameConfig) error {
	playerCell, err := NewPlayerCellConfig(raw)
	if err != nil {
		return err
	}

	gameConfig.PlayerCell = playerCell
	return nil
}

func applyCoreConfig(raw RawConfig, gameConfig *GameConfig) error {
	core, err := NewCoreConfig(raw)
	if err != nil {
		return err
	}

	gameConfig.Core = core
	return nil
}

func applyNutrientConfig(raw RawConfig, gameConfig *GameConfig) error {
	nutrient, err := NewNutrientConfig(raw)
	if err != nil {
		return err
	}

	gameConfig.Nutrient = nutrient
	return nil
}

func applyWallConfig(raw RawConfig, gameConfig *GameConfig) error {
	wall, err := NewWallConfig(raw)
	if err != nil {
		return err
	}

	gameConfig.Wall = wall
	return nil
}
