package generation

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/01JAMIL/supago.git/internal/config"
	"github.com/01JAMIL/supago.git/internal/introspection"
)

var queryNameRe = regexp.MustCompile(`^[A-Z][A-Za-z0-9]*$`)
var placeholderRe = regexp.MustCompile(`^\$(\d+)$`)

type queryMethod struct {
	name        string
	tableName   string
	structName  string
	modelPkg    string
	modelType   string
	sql         string
	args        []queryArg
	returnSlice bool
}

type queryArg struct {
	name   string
	goType string
}

// GenerateQueries generates custom repository methods for the queries
// configured under generation.queries in supago.yaml.
func GenerateQueries(schema *introspection.Schema, queries config.GenerationConfig, outputDir, modulePath string) error {
	if len(queries.Queries) == 0 {
		return nil
	}

	modelsPkg := modulePath + "/" + outputDir
	pkgBase := filepath.Base(outputDir)

	tableNames := make([]string, 0, len(queries.Queries))
	for name := range queries.Queries {
		tableNames = append(tableNames, name)
	}
	sort.Strings(tableNames)

	for _, tableName := range tableNames {
		table, ok := findTable(schema, tableName)
		if !ok {
			return fmt.Errorf("configured query table %q does not exist in the database schema", tableName)
		}

		methods, err := compileTableQueries(table, queries.Queries[tableName], pkgBase)
		if err != nil {
			return err
		}

		dir := filepath.Join(outputDir, tableName)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory for %s: %w", tableName, err)
		}

		if err := writeQueryFile(filepath.Join(dir, "queries.go"), table, methods, modelsPkg); err != nil {
			return fmt.Errorf("generate queries for %s: %w", tableName, err)
		}
	}
	return nil
}

func findTable(schema *introspection.Schema, name string) (introspection.Table, bool) {
	for _, t := range schema.Tables {
		if t.Name == name {
			return t, true
		}
	}
	return introspection.Table{}, false
}

func findColumn(table introspection.Table, name string) (introspection.Column, bool) {
	for _, c := range table.Columns {
		if c.Name == name {
			return c, true
		}
	}
	return introspection.Column{}, false
}

func compileTableQueries(table introspection.Table, queries []config.QueryConfig, pkgBase string) ([]queryMethod, error) {
	structName := toPascalCase(singularize(table.Name))

	var methods []queryMethod
	seen := make(map[string]bool)
	for _, q := range queries {
		method, err := compileQuery(table, q, pkgBase, structName)
		if err != nil {
			return nil, fmt.Errorf("table %q: %w", table.Name, err)
		}
		if seen[method.name] {
			return nil, fmt.Errorf("table %q: duplicate query name %q", table.Name, q.Name)
		}
		seen[method.name] = true
		methods = append(methods, method)
	}
	return methods, nil
}

func compileQuery(table introspection.Table, q config.QueryConfig, pkgBase, structName string) (queryMethod, error) {
	if strings.TrimSpace(q.Name) == "" {
		return queryMethod{}, fmt.Errorf("query name is required")
	}
	if !queryNameRe.MatchString(q.Name) {
		return queryMethod{}, fmt.Errorf("query name %q must be a valid exported Go identifier in PascalCase (e.g. FindByEmail)", q.Name)
	}
	if len(q.Where) == 0 {
		return queryMethod{}, fmt.Errorf("query %q must define at least one where condition", q.Name)
	}

	returnSlice := false
	switch {
	case strings.HasPrefix(q.Name, "List"):
		returnSlice = true
	case strings.HasPrefix(q.Name, "Find"), strings.HasPrefix(q.Name, "Get"):
		returnSlice = false
	default:
		return queryMethod{}, fmt.Errorf("query name %q must start with Find or Get (single result) or List (multiple results) so the generated return type is determinable", q.Name)
	}

	var clauses []string
	params := make(map[int]queryArg)
	for _, clause := range q.Where {
		c := strings.TrimSpace(clause)
		if c == "" {
			return queryMethod{}, fmt.Errorf("query %q contains an empty where condition", q.Name)
		}

		fields := strings.Fields(c)
		col, ok := findColumn(table, fields[0])
		if !ok {
			return queryMethod{}, fmt.Errorf("query %q references unknown column %q in table %q", q.Name, fields[0], table.Name)
		}

		for _, f := range fields {
			m := placeholderRe.FindStringSubmatch(f)
			if m == nil {
				continue
			}
			pos, err := strconv.Atoi(m[1])
			if err != nil {
				return queryMethod{}, fmt.Errorf("query %q has invalid placeholder %q", q.Name, f)
			}
			if _, dup := params[pos]; dup {
				return queryMethod{}, fmt.Errorf("query %q uses placeholder %s more than once", q.Name, f)
			}
			params[pos] = queryArg{name: toCamelCase(col.Name), goType: baseGoType(col.DataType)}
		}
		clauses = append(clauses, c)
	}

	if len(params) > 0 {
		max := 0
		for pos := range params {
			if pos > max {
				max = pos
			}
		}
		if len(params) != max {
			return queryMethod{}, fmt.Errorf("query %q positional parameters must be sequential starting at $1", q.Name)
		}
	}

	args := make([]queryArg, 0, len(params))
	usedNames := make(map[string]int)
	for i := 1; i <= len(params); i++ {
		a := params[i]
		name := a.name
		if n, ok := usedNames[name]; ok {
			usedNames[name] = n + 1
			name = fmt.Sprintf("%s%d", name, n+1)
		} else {
			usedNames[name] = 1
		}
		args = append(args, queryArg{name: name, goType: a.goType})
	}

	sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", columnNames(table.Columns), table.Name, strings.Join(clauses, " AND "))

	return queryMethod{
		name:        q.Name,
		tableName:   table.Name,
		structName:  structName,
		modelPkg:    pkgBase,
		modelType:   pkgBase + "." + structName,
		sql:         sql,
		args:        args,
		returnSlice: returnSlice,
	}, nil
}

func writeQueryFile(path string, table introspection.Table, methods []queryMethod, modelsPkg string) error {
	var b strings.Builder
	b.WriteString("// Code generated by SupaGo. DO NOT EDIT.\n")
	b.WriteString(fmt.Sprintf("package %s\n\n", table.Name))

	var timeImports, jsonImports, otherImports []string
	for _, m := range methods {
		for _, a := range m.args {
			switch a.goType {
			case "time.Time":
				timeImports = []string{"time"}
			case "json.RawMessage":
				jsonImports = []string{"encoding/json"}
			}
		}
	}

	b.WriteString("import (\n")
	b.WriteString("\t\"context\"\n")
	b.WriteString("\t\"fmt\"\n")
	for _, imp := range jsonImports {
		b.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
	}
	for _, imp := range timeImports {
		b.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
	}
	for _, imp := range otherImports {
		b.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
	}
	b.WriteString(fmt.Sprintf("\t\"%s\"\n", modelsPkg))
	b.WriteString(")\n\n")

	colFields := columnFields(table.Columns)
	for _, m := range methods {
		writeQueryMethod(&b, m, colFields)
	}

	src, err := format.Source([]byte(b.String()))
	if err != nil {
		return fmt.Errorf("format source: %w", err)
	}
	return os.WriteFile(path, src, 0644)
}

func writeQueryMethod(b *strings.Builder, m queryMethod, colFields string) {
	params := methodParamList(m.args)
	callArgs := callArgList(m.args)
	errorVerb := "find"
	if m.returnSlice {
		errorVerb = "list"
	}

	if m.returnSlice {
		b.WriteString(fmt.Sprintf("func (r *Repository) %s(ctx context.Context%s) ([]%s, error) {\n", m.name, params, m.modelType))
		b.WriteString(fmt.Sprintf("\trows, err := r.pool.Query(ctx, %s%s)\n", strconv.Quote(m.sql), callArgs))
		b.WriteString("\tif err != nil {\n")
		b.WriteString(fmt.Sprintf("\t\treturn nil, fmt.Errorf(\"%s %s: %%w\", err)\n", errorVerb, m.tableName))
		b.WriteString("\t}\n")
		b.WriteString("\tdefer rows.Close()\n\n")
		b.WriteString(fmt.Sprintf("\tvar items []%s\n", m.modelType))
		b.WriteString("\tfor rows.Next() {\n")
		b.WriteString(fmt.Sprintf("\t\tvar item %s\n", m.modelType))
		b.WriteString(fmt.Sprintf("\t\tif err := rows.Scan(%s); err != nil {\n", colFields))
		b.WriteString(fmt.Sprintf("\t\t\treturn nil, fmt.Errorf(\"scan %s: %%w\", err)\n", singularize(m.tableName)))
		b.WriteString("\t\t}\n")
		b.WriteString("\t\titems = append(items, item)\n")
		b.WriteString("\t}\n")
		b.WriteString("\treturn items, rows.Err()\n")
		b.WriteString("}\n\n")
		return
	}

	b.WriteString(fmt.Sprintf("func (r *Repository) %s(ctx context.Context%s) (*%s, error) {\n", m.name, params, m.modelType))
	b.WriteString(fmt.Sprintf("\trow := r.pool.QueryRow(ctx, %s%s)\n", strconv.Quote(m.sql), callArgs))
	b.WriteString(fmt.Sprintf("\tvar item %s\n", m.modelType))
	b.WriteString(fmt.Sprintf("\tif err := row.Scan(%s); err != nil {\n", colFields))
	b.WriteString(fmt.Sprintf("\t\treturn nil, fmt.Errorf(\"%s %s: %%w\", err)\n", errorVerb, m.tableName))
	b.WriteString("\t}\n")
	b.WriteString("\treturn &item, nil\n")
	b.WriteString("}\n\n")
}

func methodParamList(args []queryArg) string {
	if len(args) == 0 {
		return ""
	}
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = fmt.Sprintf("%s %s", a.name, a.goType)
	}
	return ", " + strings.Join(parts, ", ")
}

func callArgList(args []queryArg) string {
	if len(args) == 0 {
		return ""
	}
	names := make([]string, len(args))
	for i, a := range args {
		names[i] = a.name
	}
	return ", " + strings.Join(names, ", ")
}
