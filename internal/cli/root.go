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

const Version = "v1.1.0"

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "supago",
		Version: Version,
		Short:   "SupaGo - Supabase toolkit for Go projects",
		Long:    "SupaGo simplifies working with Supabase in Go projects through code generation, introspection, and project scaffolding.",
	}

	cmd.AddCommand(NewInitCmd())
	cmd.AddCommand(NewInspectCmd())
	cmd.AddCommand(NewGenerateCmd())
	cmd.AddCommand(NewVersionCmd())

	return cmd
}

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of SupaGo",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("supago version", Version)
			return nil
		},
	}
}
