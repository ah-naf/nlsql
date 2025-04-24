package db

import (
	"database/sql"
	"fmt"
	"log"

	"nlsql/internal/models"

	"github.com/lib/pq"
)

// GetTableNameList returns all public‐schema table names.
func GetTableNameList(conn *sql.DB) ([]string, error) {
	rows, err := conn.Query(`
        SELECT table_name
        FROM information_schema.tables
        WHERE table_schema='public'
        ORDER BY table_name
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		names = append(names, t)
	}
	return names, nil
}

// BriefSchema returns just table names + row counts.
func BriefSchema(conn *sql.DB) ([]models.BriefSchemaItem, error) {
	tbls, err := GetTableNameList(conn)
	if err != nil {
		return nil, err
	}
	var out []models.BriefSchemaItem
	for _, tbl := range tbls {
		var cnt int
		if err := conn.QueryRow(
			fmt.Sprintf("SELECT COUNT(*) FROM %s", pq.QuoteIdentifier(tbl)),
		).Scan(&cnt); err != nil {
			log.Printf("count %s: %v", tbl, err)
			continue
		}
		out = append(out, models.BriefSchemaItem{Name: tbl, RowCount: cnt})
	}
	return out, nil
}

// LoadFullSchema loads columns, PKs, uniques, FKs and row counts.
func LoadFullSchema(conn *sql.DB) (map[string]models.TableInfo, error) {
	const colQ = `
    SELECT
      c.table_name,
      c.column_name,
      c.data_type,
      (c.is_nullable = 'YES') AS is_nullable,
      c.column_default,
      pgd.description   AS column_description,
      tbl_pgd.description AS table_description
    FROM information_schema.columns AS c
    LEFT JOIN pg_catalog.pg_stat_all_tables AS st
      ON c.table_name = st.relname
    LEFT JOIN pg_catalog.pg_description AS pgd
      ON pgd.objoid = st.relid AND pgd.objsubid = c.ordinal_position
    LEFT JOIN pg_catalog.pg_description AS tbl_pgd
      ON tbl_pgd.objoid = st.relid AND tbl_pgd.objsubid = 0
    WHERE c.table_schema = 'public'
    ORDER BY c.table_name, c.ordinal_position;
    `
	rows, err := conn.Query(colQ)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schema := make(map[string]models.TableInfo)
	for rows.Next() {
		var tbl, col, dt string
		var isNull bool
		var def, cDesc, tDesc sql.NullString
		if err := rows.Scan(&tbl, &col, &dt, &isNull, &def, &cDesc, &tDesc); err != nil {
			return nil, err
		}
		ti := schema[tbl]
		if ti.Name == "" {
			ti = models.TableInfo{Name: tbl, Columns: []models.ColumnInfo{}, Description: tDesc}
		}
		ti.Columns = append(ti.Columns, models.ColumnInfo{
			Name:          col,
			DataType:      dt,
			IsNullable:    isNull,
			DefaultValue:  def,
			IsPrimaryKey:  false,
			IsUnique:      false,
			ForeignTable:  sql.NullString{},
			ForeignColumn: sql.NullString{},
			Description:   cDesc,
		})
		schema[tbl] = ti
	}

	// Primary keys
	const pkQ = `
    SELECT tc.table_name, kcu.column_name
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu
      ON tc.constraint_schema = kcu.constraint_schema
     AND tc.constraint_name = kcu.constraint_name
    WHERE tc.constraint_type='PRIMARY KEY' AND tc.table_schema='public';
    `
	pkRows, err := conn.Query(pkQ)
	if err != nil {
		return nil, err
	}
	defer pkRows.Close()
	for pkRows.Next() {
		var tbl, col string
		if err := pkRows.Scan(&tbl, &col); err != nil {
			return nil, err
		}
		ti := schema[tbl]
		for i := range ti.Columns {
			if ti.Columns[i].Name == col {
				ti.Columns[i].IsPrimaryKey = true
				break
			}
		}
		schema[tbl] = ti
	}

	// Unique constraints
	const uqQ = `
    SELECT tc.table_name, kcu.column_name
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu
      ON tc.constraint_schema = kcu.constraint_schema
     AND tc.constraint_name = kcu.constraint_name
    WHERE tc.constraint_type='UNIQUE' AND tc.table_schema='public';
    `
	uqRows, err := conn.Query(uqQ)
	if err != nil {
		return nil, err
	}
	defer uqRows.Close()
	for uqRows.Next() {
		var tbl, col string
		if err := uqRows.Scan(&tbl, &col); err != nil {
			return nil, err
		}
		ti := schema[tbl]
		for i := range ti.Columns {
			if ti.Columns[i].Name == col {
				ti.Columns[i].IsUnique = true
				break
			}
		}
		schema[tbl] = ti
	}

	// Foreign keys
	const fkQ = `
    SELECT
      kcu.table_name,
      kcu.column_name,
      ccu.table_name AS foreign_table_name,
      ccu.column_name AS foreign_column_name
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu
      ON tc.constraint_schema = kcu.constraint_schema
     AND tc.constraint_name = kcu.constraint_name
    JOIN information_schema.constraint_column_usage ccu
      ON tc.constraint_schema = ccu.constraint_schema
     AND tc.constraint_name = ccu.constraint_name
    WHERE tc.constraint_type='FOREIGN KEY' AND tc.table_schema='public';
    `
	fkRows, err := conn.Query(fkQ)
	if err != nil {
		return nil, err
	}
	defer fkRows.Close()
	for fkRows.Next() {
		var tbl, col, ftbl, fcol string
		if err := fkRows.Scan(&tbl, &col, &ftbl, &fcol); err != nil {
			return nil, err
		}
		ti := schema[tbl]
		for i := range ti.Columns {
			if ti.Columns[i].Name == col {
				ti.Columns[i].ForeignTable = sql.NullString{String: ftbl, Valid: true}
				ti.Columns[i].ForeignColumn = sql.NullString{String: fcol, Valid: true}
				break
			}
		}
		schema[tbl] = ti
	}

	// Row counts
	for tbl, ti := range schema {
		var cnt int
		if err := conn.QueryRow(
			fmt.Sprintf("SELECT COUNT(*) FROM %s", pq.QuoteIdentifier(tbl)),
		).Scan(&cnt); err != nil {
			log.Printf("row count %s: %v", tbl, err)
		}
		ti.RowCount = cnt
		schema[tbl] = ti
	}

	return schema, nil
}
