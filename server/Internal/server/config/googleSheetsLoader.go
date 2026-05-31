package config

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/sheets/v4"
)

func LoadKeyValueConfig(
	ctx context.Context,
	srv *sheets.Service,
	spreadsheetID string,
	sheetName string) (RawConfig, error) {
	var readRange = fmt.Sprintf("'%s'!A2:C", sheetName)

	var resp, err = srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Context(ctx).Do()

	if err != nil {
		return nil, fmt.Errorf("read sheet %s error: %w", sheetName, err)
	}

	var result = make(RawConfig)

	for i, row := range resp.Values {
		var rowNumber = i + 2

		if len(row) < 3 {
			return nil, fmt.Errorf("%s row %d: expected KEY, NAME, Value", sheetName, rowNumber)
		}

		var key = strings.TrimSpace(fmt.Sprint(row[0]))
		var value = strings.TrimSpace(fmt.Sprint(row[2]))

		if key == "" {
			return nil, fmt.Errorf("%s row %d: empty KEY", sheetName, rowNumber)
		}

		if _, exists := result[ConfigKey(key)]; exists {
			return nil, fmt.Errorf("%s row %d: duplicate KEY %q", sheetName, rowNumber, key)
		}

		result[ConfigKey(key)] = value
	}

	return result, nil
}
