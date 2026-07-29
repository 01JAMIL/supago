package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supago",
		Short: "SupaGo - Supabase toolkit for Go projects",
		Long:  "SupaGo simplifies working with Supabase in Go projects through code generation, introspection, and project scaffolding.",
	}

	cmd.AddCommand(NewInitCmd())
	cmd.AddCommand(NewInspectCmd())

	return cmd
}
