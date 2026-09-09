package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWithQueries(t *testing.T) {
	cfg := loadTempConfig(t, `
generation:
  output: internal/adapters/supago
  queries:
    users:
      - name: FindByEmail
        where:
          - email = $1
      - name: FindByName
        where:
          - full_name = $1
`)

	if cfg.Generation.Output != "internal/adapters/supago" {
		t.Fatalf("expected output internal/adapters/supago, got %q", cfg.Generation.Output)
	}
	queries, ok := cfg.Generation.Queries["users"]
	if !ok {
		t.Fatal("expected queries for table users")
	}
	if len(queries) != 2 {
		t.Fatalf("expected 2 queries, got %d", len(queries))
	}
	if queries[0].Name != "FindByEmail" {
		t.Fatalf("expected query name FindByEmail, got %q", queries[0].Name)
	}
	if len(queries[0].Where) != 1 || queries[0].Where[0] != "email = $1" {
		t.Fatalf("unexpected where conditions: %#v", queries[0].Where)
	}
	if queries[1].Name != "FindByName" {
		t.Fatalf("expected query name FindByName, got %q", queries[1].Name)
	}
}

func TestLoadMultipleTables(t *testing.T) {
	cfg := loadTempConfig(t, `
generation:
  queries:
    users:
      - name: FindByEmail
        where:
          - email = $1
    posts:
      - name: FindByAuthor
        where:
          - author_id = $1
`)

	if len(cfg.Generation.Queries) != 2 {
		t.Fatalf("expected 2 tables with queries, got %d", len(cfg.Generation.Queries))
	}
	if len(cfg.Generation.Queries["posts"]) != 1 {
		t.Fatalf("expected 1 query for posts, got %d", len(cfg.Generation.Queries["posts"]))
	}
}

func TestLoadWithoutQueries(t *testing.T) {
	cfg := loadTempConfig(t, `
generation:
  output: internal/adapters/supago
`)

	if cfg.Generation.Queries != nil {
		t.Fatalf("expected no queries, got %#v", cfg.Generation.Queries)
	}
}

func loadTempConfig(t *testing.T, content string) *Config {
	t.Helper()
	path := filepath.Join(t.TempDir(), "supago.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}
