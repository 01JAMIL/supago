package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateQueryCmdNoQueriesConfigured(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "supago.yaml"), []byte(`
generation:
  output: internal/adapters/supago
`), 0644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	cmd := newGenerateQueryCmd()
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error when queries are not configured, got %v", err)
	}
}

func TestGenerateCmdHasQuerySubcommand(t *testing.T) {
	root := NewGenerateCmd()
	sub, _, err := root.Find([]string{"query"})
	if err != nil {
		t.Fatalf("expected query subcommand to exist, got error: %v", err)
	}
	if sub.Use != "query" {
		t.Fatalf("expected query subcommand, got %q", sub.Use)
	}
}
