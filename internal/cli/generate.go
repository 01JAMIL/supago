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

func NewGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate code from the database schema",
	}

	cmd.AddCommand(newGenerateModelCmd())

	return cmd
}

func newGenerateModelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "model",
		Short: "Generate Go structs from database tables",
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

			if err := generation.GenerateModels(schema, outputDir); err != nil {
				return fmt.Errorf("generate models: %w", err)
			}

			fmt.Printf("✔ Models generated in %s/models/\n", outputDir)

			return nil
		},
	}
}
