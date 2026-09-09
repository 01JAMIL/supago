package generation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/01JAMIL/supago.git/internal/config"
	"github.com/01JAMIL/supago.git/internal/introspection"
)

const testModulePath = "github.com/example/project"

func testSchema() *introspection.Schema {
	return &introspection.Schema{
		Tables: []introspection.Table{
			{
				Name: "users",
				Columns: []introspection.Column{
					{Name: "id", DataType: "bigint", PrimaryKey: true},
					{Name: "full_name", DataType: "character varying"},
					{Name: "email", DataType: "character varying"},
					{Name: "created_at", DataType: "timestamp with time zone"},
				},
			},
			{
				Name: "posts",
				Columns: []introspection.Column{
					{Name: "id", DataType: "bigint", PrimaryKey: true},
					{Name: "author_id", DataType: "bigint"},
					{Name: "title", DataType: "character varying"},
				},
			},
		},
	}
}

func generateQueries(t *testing.T, queries config.GenerationConfig) string {
	t.Helper()
	outputDir := filepath.Join(t.TempDir(), "supago")
	if err := GenerateQueries(testSchema(), queries, outputDir, testModulePath); err != nil {
		t.Fatal(err)
	}
	return outputDir
}

func readQueryFile(t *testing.T, outputDir, table string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(outputDir, table, "queries.go"))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func TestGenerateQueriesSingle(t *testing.T) {
	dir := generateQueries(t, config.GenerationConfig{
		Queries: map[string][]config.QueryConfig{
			"users": {
				{Name: "FindByEmail", Where: []string{"email = $1"}},
			},
		},
	})

	content := readQueryFile(t, dir, "users")

	if !strings.Contains(content, "func (r *Repository) FindByEmail(ctx context.Context, email string) (*supago.User, error) {") {
		t.Errorf("expected FindByEmail signature with email parameter, got:\n%s", content)
	}
	if !strings.Contains(content, `"SELECT id, full_name, email, created_at FROM users WHERE email = $1", email`) {
		t.Errorf("expected SQL with email = $1 and the email argument, got:\n%s", content)
	}
	if !strings.Contains(content, `&item.ID, &item.FullName, &item.Email, &item.CreatedAt`) {
		t.Errorf("expected scan of all columns, got:\n%s", content)
	}
}

func TestGenerateQueriesMultipleTables(t *testing.T) {
	dir := generateQueries(t, config.GenerationConfig{
		Queries: map[string][]config.QueryConfig{
			"users": {
				{Name: "FindByEmail", Where: []string{"email = $1"}},
			},
			"posts": {
				{Name: "ListByAuthor", Where: []string{"author_id = $1"}},
			},
		},
	})

	users := readQueryFile(t, dir, "users")
	if !strings.Contains(users, "func (r *Repository) FindByEmail(ctx context.Context, email string) (*supago.User, error) {") {
		t.Errorf("expected users FindByEmail, got:\n%s", users)
	}

	posts := readQueryFile(t, dir, "posts")
	if !strings.Contains(posts, "func (r *Repository) ListByAuthor(ctx context.Context, authorID int64) ([]supago.Post, error) {") {
		t.Errorf("expected posts ListByAuthor, got:\n%s", posts)
	}
}

func TestGenerateQueriesMultipleConditions(t *testing.T) {
	dir := generateQueries(t, config.GenerationConfig{
		Queries: map[string][]config.QueryConfig{
			"users": {
				{Name: "FindByEmailAndName", Where: []string{"email = $1", "full_name = $2"}},
			},
		},
	})

	content := readQueryFile(t, dir, "users")
	if !strings.Contains(content, "func (r *Repository) FindByEmailAndName(ctx context.Context, email string, fullName string) (*supago.User, error) {") {
		t.Errorf("expected method with two typed parameters, got:\n%s", content)
	}
	if !strings.Contains(content, `WHERE email = $1 AND full_name = $2", email, fullName`) {
		t.Errorf("expected combined WHERE clause with both arguments, got:\n%s", content)
	}
}

func TestGenerateQueriesListSingleResultConvention(t *testing.T) {
	dir := generateQueries(t, config.GenerationConfig{
		Queries: map[string][]config.QueryConfig{
			"users": {
				{Name: "ListByEmail", Where: []string{"email = $1"}},
			},
		},
	})

	content := readQueryFile(t, dir, "users")
	if !strings.Contains(content, "func (r *Repository) ListByEmail(ctx context.Context, email string) ([]supago.User, error) {") {
		t.Errorf("expected ListByEmail to return a slice, got:\n%s", content)
	}
}

func TestGenerateQueriesGetSingleResultConvention(t *testing.T) {
	dir := generateQueries(t, config.GenerationConfig{
		Queries: map[string][]config.QueryConfig{
			"users": {
				{Name: "GetByEmail", Where: []string{"email = $1"}},
			},
		},
	})

	content := readQueryFile(t, dir, "users")
	if !strings.Contains(content, "func (r *Repository) GetByEmail(ctx context.Context, email string) (*supago.User, error) {") {
		t.Errorf("expected GetByEmail to return a single item, got:\n%s", content)
	}
}

func TestGenerateQueriesTimeColumnParameter(t *testing.T) {
	dir := generateQueries(t, config.GenerationConfig{
		Queries: map[string][]config.QueryConfig{
			"users": {
				{Name: "FindByCreatedAt", Where: []string{"created_at = $1"}},
			},
		},
	})

	content := readQueryFile(t, dir, "users")
	if !strings.Contains(content, "func (r *Repository) FindByCreatedAt(ctx context.Context, createdAt time.Time) (*supago.User, error) {") {
		t.Errorf("expected time.Time parameter, got:\n%s", content)
	}
	if !strings.Contains(content, `"time"`) {
		t.Errorf("expected time import in generated file, got:\n%s", content)
	}
}

func TestGenerateQueriesUnknownTable(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "supago")
	err := GenerateQueries(testSchema(), config.GenerationConfig{
		Queries: map[string][]config.QueryConfig{
			"missing": {
				{Name: "FindAll", Where: []string{"id = $1"}},
			},
		},
	}, outputDir, testModulePath)
	if err == nil {
		t.Fatal("expected error for unknown table")
	}
	if !strings.Contains(err.Error(), `"missing"`) {
		t.Errorf("expected error to mention the unknown table, got %v", err)
	}
}

func TestGenerateQueriesMissingName(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "supago")
	err := GenerateQueries(testSchema(), config.GenerationConfig{
		Queries: map[string][]config.QueryConfig{
			"users": {
				{Name: "", Where: []string{"email = $1"}},
			},
		},
	}, outputDir, testModulePath)
	if err == nil {
		t.Fatal("expected error for missing query name")
	}
	if !strings.Contains(err.Error(), "query name is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGenerateQueriesInvalidName(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "supago")
	err := GenerateQueries(testSchema(), config.GenerationConfig{
		Queries: map[string][]config.QueryConfig{
			"users": {
				{Name: "findByEmail", Where: []string{"email = $1"}},
			},
		},
	}, outputDir, testModulePath)
	if err == nil {
		t.Fatal("expected error for invalid query name")
	}
}

func TestGenerateQueriesMissingPrefix(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "supago")
	err := GenerateQueries(testSchema(), config.GenerationConfig{
		Queries: map[string][]config.QueryConfig{
			"users": {
				{Name: "ByEmail", Where: []string{"email = $1"}},
			},
		},
	}, outputDir, testModulePath)
	if err == nil {
		t.Fatal("expected error for query name without a recognized prefix")
	}
}

func TestGenerateQueriesEmptyWhere(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "supago")
	err := GenerateQueries(testSchema(), config.GenerationConfig{
		Queries: map[string][]config.QueryConfig{
			"users": {
				{Name: "FindAll", Where: []string{}},
			},
		},
	}, outputDir, testModulePath)
	if err == nil {
		t.Fatal("expected error for missing where conditions")
	}
}

func TestGenerateQueriesDuplicateName(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "supago")
	err := GenerateQueries(testSchema(), config.GenerationConfig{
		Queries: map[string][]config.QueryConfig{
			"users": {
				{Name: "FindByEmail", Where: []string{"email = $1"}},
				{Name: "FindByEmail", Where: []string{"email = $1"}},
			},
		},
	}, outputDir, testModulePath)
	if err == nil {
		t.Fatal("expected error for duplicate query name")
	}
	if !strings.Contains(err.Error(), "duplicate query name") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGenerateQueriesUnknownColumn(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "supago")
	err := GenerateQueries(testSchema(), config.GenerationConfig{
		Queries: map[string][]config.QueryConfig{
			"users": {
				{Name: "FindByAge", Where: []string{"age = $1"}},
			},
		},
	}, outputDir, testModulePath)
	if err == nil {
		t.Fatal("expected error for unknown column")
	}
	msg := err.Error()
	if !strings.Contains(msg, "unknown column") || !strings.Contains(msg, "age") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGenerateQueriesNonSequentialParams(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "supago")
	err := GenerateQueries(testSchema(), config.GenerationConfig{
		Queries: map[string][]config.QueryConfig{
			"users": {
				{Name: "FindByEmail", Where: []string{"email = $2"}},
			},
		},
	}, outputDir, testModulePath)
	if err == nil {
		t.Fatal("expected error for non-sequential parameters")
	}
	if !strings.Contains(err.Error(), "sequential") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGenerateQueriesDuplicatePlaceholder(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "supago")
	err := GenerateQueries(testSchema(), config.GenerationConfig{
		Queries: map[string][]config.QueryConfig{
			"users": {
				{Name: "FindByEmail", Where: []string{"email = $1", "full_name = $1"}},
			},
		},
	}, outputDir, testModulePath)
	if err == nil {
		t.Fatal("expected error for reused placeholder")
	}
	if !strings.Contains(err.Error(), "more than once") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGenerateQueriesNoQueries(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "supago")
	if err := GenerateQueries(testSchema(), config.GenerationConfig{}, outputDir, testModulePath); err != nil {
		t.Fatal(err)
	}
}

func TestCompileQueryGeneratedOutputIsFormattedGo(t *testing.T) {
	dir := generateQueries(t, config.GenerationConfig{
		Queries: map[string][]config.QueryConfig{
			"users": {
				{Name: "FindByEmail", Where: []string{"email = $1"}},
				{Name: "FindByEmailAndName", Where: []string{"email = $1", "full_name = $2"}},
			},
			"posts": {
				{Name: "ListByAuthor", Where: []string{"author_id = $1"}},
			},
		},
	})

	for _, table := range []string{"users", "posts"} {
		content := readQueryFile(t, dir, table)
		if !strings.Contains(content, "// Code generated by SupaGo. DO NOT EDIT.") {
			t.Errorf("expected generated header in %s/queries.go", table)
		}
		if !strings.Contains(content, "package "+table) {
			t.Errorf("expected package %s in queries.go", table)
		}
	}
}
