package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/01JAMIL/supago.git/internal/config"
	"github.com/01JAMIL/supago.git/internal/database"
	"github.com/01JAMIL/supago.git/internal/generation"
	"github.com/01JAMIL/supago.git/internal/introspection"
	"github.com/spf13/cobra"
)

func newGenerateRepositoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "repository",
		Short: "Generate Go repositories from database tables",
		RunE: func(cmd *cobra.Command, args []string) error {
			PrintBanner()

			ctx := context.Background()

			cfg, err := config.Load("supago.yaml")
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			if cfg.Database.URL == "" {
				return fmt.Errorf("database URL is empty: set SUPAGO_DATABASE_URL in your environment")
			}

			pool, err := database.Connect(ctx, cfg.Database.URL)
			if err != nil {
				return err
			}
			defer pool.Close()

			fmt.Print("\n\n🔍 Inspecting database... \n\n")

			schema, err := introspection.Inspect(ctx, pool)
			if err != nil {
				return err
			}

			fmt.Printf("✔ %d tables found\n", schema.TableCount)
			fmt.Printf("✔ %d columns found\n\n", schema.ColumnCount)

			outputDir := cfg.Generation.Output
			if outputDir == "" {
				outputDir = "internal/generated"
			}

			modulePath, err := readModulePath()
			if err != nil {
				return fmt.Errorf("read go.mod: %w", err)
			}

			if err := generation.GenerateRepositories(schema, outputDir, modulePath); err != nil {
				return fmt.Errorf("generate repositories: %w", err)
			}

			fmt.Printf("✔ Repositories generated in %s/<table>/repository.go\n", outputDir)

			return nil
		},
	}
}

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
