package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const SupagoBanner = `
███████╗██╗   ██╗██████╗  █████╗  ██████╗  ██████╗
██╔════╝██║   ██║██╔══██╗██╔══██╗██╔════╝ ██╔═══██╗
███████╗██║   ██║██████╔╝███████║██║  ███╗██║   ██║
╚════██║██║   ██║██╔═══╝ ██╔══██║██║   ██║██║   ██║
███████║╚██████╔╝██║     ██║  ██║╚██████╔╝╚██████╔╝
╚══════╝ ╚═════╝ ╚═╝     ╚═╝  ╚═╝ ╚═════╝  ╚═════╝
`

func PrintBanner() {
	fmt.Print(SupagoBanner)
}

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supago",
		Short: "SupaGo - Supabase toolkit for Go projects",
		Long:  "SupaGo simplifies working with Supabase in Go projects through code generation, introspection, and project scaffolding.",
	}

	cmd.AddCommand(NewInitCmd())
	cmd.AddCommand(NewInspectCmd())
	cmd.AddCommand(NewGenerateCmd())

	return cmd
}
