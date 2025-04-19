package models

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type ColumnInfo struct {
	Name          string         `json:"name"`
	DataType      string         `json:"data_type"`
	ForeignTable  sql.NullString `json:"foreign_table,omitempty"`
	ForeignColumn sql.NullString `json:"foreign_column,omitempty"`
	IsPrimaryKey  bool           `json:"is_primary_key"`
}

// GetDatabases returns non-template database names for a given DB connection
func GetDatabases(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SELECT datname FROM pg_database WHERE datistemplate = false;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dbs []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		dbs = append(dbs, name)
	}
	return dbs, nil
}

// GetSchema fetches table/column schema for the provided DB connection
func GetSchema(db *sql.DB) (map[string][]ColumnInfo, error) {
	const query = `
SELECT
    c.table_name,
    c.column_name,
    c.data_type,
    ccu.table_name    AS foreign_table,
    ccu.column_name   AS foreign_column,
    -- this will be TRUE when the column is part of a PRIMARY KEY constraint
    CASE WHEN pk.constraint_type = 'PRIMARY KEY' THEN true ELSE false END AS is_primary_key
FROM information_schema.columns AS c

-- FK detection
LEFT JOIN information_schema.key_column_usage AS kcu_fk
  ON c.table_schema = kcu_fk.constraint_schema
  AND c.table_name   = kcu_fk.table_name
  AND c.column_name  = kcu_fk.column_name
LEFT JOIN information_schema.table_constraints AS tc
  ON kcu_fk.constraint_schema = tc.constraint_schema
  AND kcu_fk.constraint_name   = tc.constraint_name
  AND tc.constraint_type      = 'FOREIGN KEY'
LEFT JOIN information_schema.constraint_column_usage AS ccu
  ON tc.constraint_schema = ccu.constraint_schema
  AND tc.constraint_name   = ccu.constraint_name

-- PK detection
LEFT JOIN information_schema.key_column_usage AS kcu_pk
  ON c.table_schema = kcu_pk.constraint_schema
  AND c.table_name   = kcu_pk.table_name
  AND c.column_name  = kcu_pk.column_name
LEFT JOIN information_schema.table_constraints AS pk
  ON kcu_pk.constraint_schema = pk.constraint_schema
  AND kcu_pk.constraint_name   = pk.constraint_name
  AND pk.constraint_type      = 'PRIMARY KEY'

WHERE c.table_schema = 'public'
ORDER BY c.table_name, c.ordinal_position;
`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schema := make(map[string][]ColumnInfo)
	for rows.Next() {
		var (
			tbl      string
			col      string
			typ      string
			fkTable  sql.NullString
			fkColumn sql.NullString
			isPK     bool
		)
		if err := rows.Scan(&tbl, &col, &typ, &fkTable, &fkColumn, &isPK); err != nil {
			return nil, err
		}

		schema[tbl] = append(schema[tbl], ColumnInfo{
			Name:          col,
			DataType:      typ,
			ForeignTable:  fkTable,
			ForeignColumn: fkColumn,
			IsPrimaryKey:  isPK,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return schema, nil
}
