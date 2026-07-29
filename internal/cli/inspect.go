package cli

import (
	"context"
	"fmt"

	"github.com/01JAMIL/supago.git/internal/config"
	"github.com/01JAMIL/supago.git/internal/database"
	"github.com/01JAMIL/supago.git/internal/introspection"
	"github.com/spf13/cobra"
)

func NewInspectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect",
		Short: "Inspect the database schema",
		Long:  "Connects to the PostgreSQL database and displays all tables with their columns, types, nullability, and primary keys.",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			schema, err := introspection.Inspect(ctx, pool)
			if err != nil {
				return err
			}

			for _, t := range schema.Tables {
				fmt.Printf("Table: %s\n", t.Name)
				for _, c := range t.Columns {
					nullable := "NOT NULL "
					if c.Nullable {
						nullable = "NULLABLE"
					}
					pk := ""
					if c.PrimaryKey {
						pk = " PK"
					}
					fmt.Printf("  %-20s %-25s %s%s\n", c.Name, c.DataType, nullable, pk)
				}
				fmt.Println()
			}

			return nil
		},
	}
}
