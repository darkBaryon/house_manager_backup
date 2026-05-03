package configpath

import (
	"fmt"
	"os"
)

const DefaultConfig = "config/config.local.yaml"

func Resolve(cliPath string) (string, error) {
	path := DefaultConfig
	if cliPath != "" {
		path = cliPath
	}

	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("config file not found: %s", path)
	}
	return path, nil
}
