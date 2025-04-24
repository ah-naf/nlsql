package models

import "database/sql"

// ColumnInfo describes one column.
type ColumnInfo struct {
	Name          string         `json:"name"`
	DataType      string         `json:"data_type"`
	IsNullable    bool           `json:"is_nullable"`
	DefaultValue  sql.NullString `json:"default_value"`
	IsPrimaryKey  bool           `json:"is_primary_key"`
	IsUnique      bool           `json:"is_unique"`
	ForeignTable  sql.NullString `json:"foreign_table"`
	ForeignColumn sql.NullString `json:"foreign_column"`
	Description   sql.NullString `json:"description"`
}

// TableInfo describes a full table.
type TableInfo struct {
	Name        string         `json:"name"`
	Columns     []ColumnInfo   `json:"columns"`
	Description sql.NullString `json:"description"`
	RowCount    int            `json:"row_count"`
}
