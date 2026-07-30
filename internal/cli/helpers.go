package cli

import (
	"fmt"
	"os"
	"strings"
)

func readModulePath() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}
	lines := strings.SplitN(string(data), "\n", 2)
	if len(lines) == 0 {
		return "", fmt.Errorf("invalid go.mod")
	}
	parts := strings.Fields(lines[0])
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid go.mod: missing module path")
	}
	return parts[1], nil
}
