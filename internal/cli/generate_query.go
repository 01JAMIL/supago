package cli

import (
	"context"
	"fmt"

	"github.com/01JAMIL/supago.git/internal/config"
	"github.com/01JAMIL/supago.git/internal/database"
	"github.com/01JAMIL/supago.git/internal/generation"
	"github.com/01JAMIL/supago.git/internal/introspection"
	"github.com/spf13/cobra"
)

func newGenerateQueryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "query",
		Short: "Generate custom repository queries from supago.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			PrintBanner()

			ctx := context.Background()

			cfg, err := config.Load("supago.yaml")
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			if len(cfg.Generation.Queries) == 0 {
				fmt.Println()
				fmt.Println("ℹ  No custom queries configured.")
				fmt.Println("ℹ  Add a `queries` section under `generation` in supago.yaml to generate custom repository methods.")
				fmt.Println()
				return nil
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

			if err := generation.GenerateQueries(schema, cfg.Generation, outputDir, modulePath); err != nil {
				return fmt.Errorf("generate queries: %w", err)
			}

			fmt.Printf("✔ %d custom queries generated in %s/<table>/queries.go\n", countQueries(cfg), outputDir)
			fmt.Println("ℹ  The generated methods extend the existing repository for each table.")
			fmt.Println("ℹ  Run `supago generate repository` first if the repositories do not exist yet.")

			return nil
		},
	}
}

func countQueries(cfg *config.Config) int {
	count := 0
	for _, queries := range cfg.Generation.Queries {
		count += len(queries)
	}
	return count
}
