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

func newGenerateAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "all",
		Short: "Run the full generation pipeline (model, repository, crud)",
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

			modulePath, err := readModulePath()
			if err != nil {
				return fmt.Errorf("read go.mod: %w", err)
			}

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

			fmt.Println("⚙ Generating models...")
			if err := generation.GenerateModels(schema, outputDir); err != nil {
				return fmt.Errorf("generate models: %w", err)
			}
			fmt.Println("✔ Models generated")
			fmt.Println()

			fmt.Println("⚙ Generating repositories...")
			if err := generation.GenerateRepositories(schema, outputDir, modulePath); err != nil {
				return fmt.Errorf("generate repositories: %w", err)
			}
			fmt.Println("✔ Repositories generated")
			fmt.Println()

			fmt.Println("⚙ Generating CRUD...")
			if err := generation.GenerateCRUD(schema, modulePath); err != nil {
				return fmt.Errorf("generate crud: %w", err)
			}
			fmt.Println("✔ Services generated")
			fmt.Println("✔ Handlers generated")
			fmt.Println()

			fmt.Println("🎉 SupaGo generation completed successfully!")

			return nil
		},
	}
}
