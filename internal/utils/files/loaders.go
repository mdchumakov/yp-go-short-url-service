package files

import (
	"encoding/json"
	"fmt"
	"os"
)

func ParseJSON[T any](filePath string) (*T, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var parsedJSON T
	if err = json.Unmarshal(b, &parsedJSON); err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	return &parsedJSON, nil
}
