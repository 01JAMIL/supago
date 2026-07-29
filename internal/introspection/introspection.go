package introspection

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Column struct {
	Name       string
	DataType   string
	Nullable   bool
	PrimaryKey bool
}

type Table struct {
	Name    string
	Columns []Column
}

type Schema struct {
	Tables []Table
}

func Inspect(ctx context.Context, pool *pgxpool.Pool) (*Schema, error) {
	rows, err := pool.Query(ctx, `
		SELECT
			t.table_name,
			c.column_name,
			c.data_type,
			c.is_nullable,
			CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END AS is_primary_key
		FROM information_schema.tables t
		JOIN information_schema.columns c
			ON c.table_schema = t.table_schema AND c.table_name = t.table_name
		LEFT JOIN (
			SELECT kcu.column_name, kcu.table_name, kcu.table_schema
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu
				ON kcu.constraint_name = tc.constraint_name
				AND kcu.table_schema = tc.table_schema
			WHERE tc.constraint_type = 'PRIMARY KEY'
		) pk
			ON pk.table_schema = c.table_schema
			AND pk.table_name = c.table_name
			AND pk.column_name = c.column_name
		WHERE t.table_schema = 'public'
			AND t.table_type = 'BASE TABLE'
		ORDER BY t.table_name, c.ordinal_position
	`)
	if err != nil {
		return nil, fmt.Errorf("query schema: %w", err)
	}
	defer rows.Close()

	tables := make(map[string][]Column)
	var order []string

	for rows.Next() {
		var tableName, colName, dataType, nullable string
		var isPK bool
		if err := rows.Scan(&tableName, &colName, &dataType, &nullable, &isPK); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		if _, ok := tables[tableName]; !ok {
			order = append(order, tableName)
		}
		tables[tableName] = append(tables[tableName], Column{
			Name:       colName,
			DataType:   dataType,
			Nullable:   nullable == "YES",
			PrimaryKey: isPK,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	schema := &Schema{}
	for _, name := range order {
		schema.Tables = append(schema.Tables, Table{
			Name:    name,
			Columns: tables[name],
		})
	}

	return schema, nil
}
