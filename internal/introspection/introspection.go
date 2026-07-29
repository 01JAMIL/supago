package introspection

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ForeignKeyRef struct {
	Table  string
	Column string
}

type Column struct {
	Name       string
	DataType   string
	Nullable   bool
	PrimaryKey bool
	ForeignKey *ForeignKeyRef
}

type Table struct {
	Name    string
	Columns []Column
}

type Schema struct {
	Tables      []Table
	TableCount  int
	ColumnCount int
}

func Inspect(ctx context.Context, pool *pgxpool.Pool) (*Schema, error) {
	rows, err := pool.Query(ctx, `
		SELECT
			t.table_name,
			c.column_name,
			c.data_type,
			c.is_nullable,
			CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END AS is_primary_key,
			fk.foreign_table_name,
			fk.foreign_column_name
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
		LEFT JOIN (
			SELECT
				kcu.column_name,
				kcu.table_name,
				kcu.table_schema,
				ccu.table_name AS foreign_table_name,
				ccu.column_name AS foreign_column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu
				ON kcu.constraint_name = tc.constraint_name
				AND kcu.table_schema = tc.table_schema
			JOIN information_schema.constraint_column_usage ccu
				ON ccu.constraint_name = tc.constraint_name
				AND ccu.table_schema = tc.table_schema
			WHERE tc.constraint_type = 'FOREIGN KEY'
		) fk
			ON fk.table_schema = c.table_schema
			AND fk.table_name = c.table_name
			AND fk.column_name = c.column_name
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
	var totalCols int

	for rows.Next() {
		var tableName, colName, dataType, nullable string
		var isPK bool
		var fkTable, fkCol *string
		if err := rows.Scan(&tableName, &colName, &dataType, &nullable, &isPK, &fkTable, &fkCol); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		if _, ok := tables[tableName]; !ok {
			order = append(order, tableName)
		}

		var fk *ForeignKeyRef
		if fkTable != nil && fkCol != nil {
			fk = &ForeignKeyRef{Table: *fkTable, Column: *fkCol}
		}

		tables[tableName] = append(tables[tableName], Column{
			Name:       colName,
			DataType:   dataType,
			Nullable:   nullable == "YES",
			PrimaryKey: isPK,
			ForeignKey: fk,
		})
		totalCols++
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	schema := &Schema{
		TableCount:  len(order),
		ColumnCount: totalCols,
	}
	for _, name := range order {
		schema.Tables = append(schema.Tables, Table{
			Name:    name,
			Columns: tables[name],
		})
	}

	return schema, nil
}
