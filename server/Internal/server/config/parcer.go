package config

import (
	"fmt"
	"strconv"
)

type RawConfig map[string]string

func (c RawConfig) Int(key string) (int, error) {
	value, err := c.String(key)

	if err != nil {
		return 0, err
	}

	parsed, err := strconv.Atoi(value)

	if err != nil {
		return 0, fmt.Errorf("failed to convert %s to int: %s", value, err)
	}

	return parsed, nil
}

func (c RawConfig) Float(key string) (float64, error) {
	value, err := c.String(key)

	if err != nil {
		return 0, err
	}

	parsed, err := strconv.ParseFloat(value, 64)

	if err != nil {
		return 0, fmt.Errorf("failed to convert %s to float: %s", value, err)
	}

	return parsed, nil
}

func (c RawConfig) Bool(key string) (bool, error) {
	value, err := c.String(key)

	if err != nil {
		return false, err
	}

	parsed, err := strconv.ParseBool(value)

	if err != nil {
		return false, fmt.Errorf("failed to convert %s to bool: %s", value, err)
	}

	return parsed, nil
}

func (c RawConfig) String(key string) (string, error) {
	var value, ok = c[key]

	if !ok {
		return "", fmt.Errorf("key %s not found", key)
	}

	return value, nil
}
