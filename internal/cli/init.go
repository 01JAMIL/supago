package cli

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed templates/supago.yaml
var supagoYamlContent string

//go:embed templates/env.example
var envExampleContent string

func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize SupaGo in an existing Go project",
		Long: `Initialize SupaGo in the current directory.

Verifies that the current directory contains a go.mod file, creates a supago.yaml
configuration file, and generates a .env.example with the required environment variables.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
				return fmt.Errorf("go.mod not found: run this command from the root of a Go project")
			}

			reader := bufio.NewReader(os.Stdin)

			if err := writeWithConfirm(reader, "supago.yaml", supagoYamlContent); err != nil {
				return err
			}

			if err := writeWithConfirm(reader, ".env.example", envExampleContent); err != nil {
				return err
			}

			fmt.Println("SupaGo initialized successfully.")
			return nil
		},
	}
}

func writeWithConfirm(r *bufio.Reader, name, content string) error {
	if _, err := os.Stat(name); err == nil {
		fmt.Printf("%s already exists. Overwrite? [y/N]: ", name)
		resp, _ := r.ReadString('\n')
		resp = strings.TrimSpace(strings.ToLower(resp))
		if resp != "y" && resp != "yes" {
			fmt.Printf("Skipping %s.\n", name)
			return nil
		}
	}
	return os.WriteFile(name, []byte(content), 0644)
}
