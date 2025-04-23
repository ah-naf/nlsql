package models

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

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

type TableInfo struct {
	Name        string         `json:"name"`
	Columns     []ColumnInfo   `json:"columns"`
	Description sql.NullString `json:"description"`
	RowCount    int            `json:"row_count"`
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
func GetSchema(db *sql.DB) (map[string]TableInfo, error) {
	query := `
    WITH unique_constraints AS (
        SELECT tc.table_name, kcu.column_name
        FROM information_schema.table_constraints AS tc
        JOIN information_schema.key_column_usage AS kcu
            ON tc.constraint_schema = kcu.constraint_schema
            AND tc.constraint_name = kcu.constraint_name
        WHERE tc.constraint_type = 'UNIQUE'
        AND tc.table_schema = 'public'
    )
    SELECT
        c.table_name,
        c.column_name,
        c.data_type,
        c.is_nullable = 'YES' AS is_nullable,
        c.column_default,
        -- FK detection
        ccu.table_name AS foreign_table,
        ccu.column_name AS foreign_column,
        -- PK detection
        CASE WHEN pk.constraint_type = 'PRIMARY KEY' THEN true ELSE false END AS is_primary_key,
        -- UNIQUE detection
        CASE WHEN uc.column_name IS NOT NULL THEN true ELSE false END AS is_unique,
        -- Column description
        pgd.description AS column_description,
        -- Table description (will be the same for all columns in a table)
        tbl_pgd.description AS table_description
    FROM information_schema.columns AS c
    -- FK detection
    LEFT JOIN information_schema.key_column_usage AS kcu_fk
        ON c.table_schema = kcu_fk.constraint_schema
        AND c.table_name = kcu_fk.table_name
        AND c.column_name = kcu_fk.column_name
    LEFT JOIN information_schema.table_constraints AS tc
        ON kcu_fk.constraint_schema = tc.constraint_schema
        AND kcu_fk.constraint_name = tc.constraint_name
        AND tc.constraint_type = 'FOREIGN KEY'
    LEFT JOIN information_schema.constraint_column_usage AS ccu
        ON tc.constraint_schema = ccu.constraint_schema
        AND tc.constraint_name = ccu.constraint_name
    -- PK detection
    LEFT JOIN information_schema.key_column_usage AS kcu_pk
        ON c.table_schema = kcu_pk.constraint_schema
        AND c.table_name = kcu_pk.table_name
        AND c.column_name = kcu_pk.column_name
    LEFT JOIN information_schema.table_constraints AS pk
        ON kcu_pk.constraint_schema = pk.constraint_schema
        AND kcu_pk.constraint_name = pk.constraint_name
        AND pk.constraint_type = 'PRIMARY KEY'
    -- UNIQUE detection
    LEFT JOIN unique_constraints AS uc
        ON c.table_name = uc.table_name
        AND c.column_name = uc.column_name
    -- Column comments
    LEFT JOIN pg_catalog.pg_stat_all_tables AS st
        ON c.table_name = st.relname
    LEFT JOIN pg_catalog.pg_description AS pgd
        ON pgd.objoid = st.relid
        AND pgd.objsubid = c.ordinal_position
    -- Table comments
    LEFT JOIN pg_catalog.pg_description AS tbl_pgd
        ON tbl_pgd.objoid = st.relid
        AND tbl_pgd.objsubid = 0
    WHERE c.table_schema = 'public'
    ORDER BY c.table_name, c.ordinal_position;
    `

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]TableInfo)

	for rows.Next() {
		var (
			tableName     string
			columnName    string
			dataType      string
			isNullable    bool
			defaultValue  sql.NullString
			foreignTable  sql.NullString
			foreignColumn sql.NullString
			isPrimaryKey  bool
			isUnique      bool
			columnDesc    sql.NullString
			tableDesc     sql.NullString
		)

		if err := rows.Scan(
			&tableName, &columnName, &dataType, &isNullable, &defaultValue,
			&foreignTable, &foreignColumn, &isPrimaryKey, &isUnique,
			&columnDesc, &tableDesc,
		); err != nil {
			return nil, err
		}

		col := ColumnInfo{
			Name:          columnName,
			DataType:      dataType,
			IsNullable:    isNullable,
			DefaultValue:  defaultValue,
			IsPrimaryKey:  isPrimaryKey,
			IsUnique:      isUnique,
			ForeignTable:  foreignTable,
			ForeignColumn: foreignColumn,
			Description:   columnDesc,
		}

		tableInfo, exists := result[tableName]
		if !exists {
			tableInfo = TableInfo{
				Name:        tableName,
				Description: tableDesc,
				Columns:     []ColumnInfo{},
			}
		}

		tableInfo.Columns = append(tableInfo.Columns, col)
		result[tableName] = tableInfo
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Get row counts for each table
	for tableName, info := range result {
		var count int
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", pq.QuoteIdentifier(tableName))).Scan(&count)
		if err != nil {
			// Log the error but continue with the process
			log.Printf("Failed to get row count for table %s: %v", tableName, err)
			continue
		}

		info.RowCount = count
		result[tableName] = info
	}

	return result, nil
}
