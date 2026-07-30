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

func newGenerateCrudCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "crud",
		Short: "Generate CRUD interfaces, services, and handlers from database tables",
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

			modulePath, err := readModulePath()
			if err != nil {
				return fmt.Errorf("read go.mod: %w", err)
			}

			if err := generation.GenerateCRUD(schema, modulePath); err != nil {
				return fmt.Errorf("generate crud: %w", err)
			}

			fmt.Println("✔ CRUD generated in internal/<table>/")
			return nil
		},
	}
}
