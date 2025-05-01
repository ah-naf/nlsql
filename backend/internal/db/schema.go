// db/schema.go
package db

import (
	"database/sql"
	"fmt"
	"log"

	"nlsql/internal/models"

	"github.com/lib/pq"
)

// BriefSchema returns just table names + row counts.
func BriefSchema(conn *sql.DB, provider string) ([]models.BriefSchemaItem, error) {
	tbls, err := GetTableNameList(conn, provider)
	if err != nil {
		return nil, err
	}
	var out []models.BriefSchemaItem
	for _, tbl := range tbls {
		var cnt int
		var query string

		switch provider {
		case "postgres", "postgresql":
			query = fmt.Sprintf("SELECT COUNT(*) FROM %s", pq.QuoteIdentifier(tbl))
		case "mysql":
			query = fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tbl)
		case "demo", "sqlite", "sqlite3":
			query = fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, tbl)
		default:
			return nil, fmt.Errorf("unsupported provider: %s", provider)
		}

		if err := conn.QueryRow(query).Scan(&cnt); err != nil {
			log.Printf("count %s: %v", tbl, err)
			continue
		}
		out = append(out, models.BriefSchemaItem{Name: tbl, RowCount: cnt})
	}
	return out, nil
}

// GetTableNameList returns all table names for the given provider.
func GetTableNameList(conn *sql.DB, provider string) ([]string, error) {
	var query string

	switch provider {
	case "postgres", "postgresql":
		query = `
			SELECT table_name
			FROM information_schema.tables
			WHERE table_schema = 'public'
			  AND table_type = 'BASE TABLE'
			ORDER BY table_name
		`
	case "mysql":
		query = `
			SELECT table_name
			FROM information_schema.tables
			WHERE table_schema = DATABASE()
			  AND table_type = 'BASE TABLE'
			ORDER BY table_name
		`
	case "demo", "sqlite", "sqlite3":
		query = `
			SELECT name
			FROM sqlite_master
			WHERE type='table'
			  AND name NOT LIKE 'sqlite_%'
			ORDER BY name
		`
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	rows, err := conn.Query(query)
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

// LoadFullSchema loads columns, PKs, uniques, FKs and row counts.
func LoadFullSchema(conn *sql.DB, provider string) (map[string]models.TableInfo, error) {
	// Handle in-memory SQLite/demo separately
	if provider == "demo" || provider == "sqlite" || provider == "sqlite3" {
		schema := make(map[string]models.TableInfo)

		// 1) List tables
		rows, err := conn.Query(`
			SELECT name
			FROM sqlite_master
			WHERE type='table'
			  AND name NOT LIKE 'sqlite_%'
			ORDER BY name;
		`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var tables []string
		for rows.Next() {
			var tbl string
			if err := rows.Scan(&tbl); err != nil {
				return nil, err
			}
			tables = append(tables, tbl)
		}

		// 2) For each table, gather details
		for _, tbl := range tables {
			ti := models.TableInfo{
				Name:        tbl,
				Columns:     []models.ColumnInfo{},
				Description: sql.NullString{},
			}

			// Columns
			colRows, err := conn.Query(fmt.Sprintf(`PRAGMA table_info("%s")`, tbl))
			if err != nil {
				return nil, err
			}
			for colRows.Next() {
				var cid, notnull, pk int
				var name, ctype string
				var dfltValue sql.NullString
				if err := colRows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
					colRows.Close()
					return nil, err
				}
				ti.Columns = append(ti.Columns, models.ColumnInfo{
					Name:          name,
					DataType:      ctype,
					IsNullable:    notnull == 0,
					DefaultValue:  dfltValue,
					IsPrimaryKey:  pk > 0,
					IsUnique:      false,
					ForeignTable:  sql.NullString{},
					ForeignColumn: sql.NullString{},
					Description:   sql.NullString{},
				})
			}
			colRows.Close()

			// Unique constraints via indexes
			idxRows, err := conn.Query(fmt.Sprintf(`PRAGMA index_list("%s")`, tbl))
			if err == nil {
				for idxRows.Next() {
					var seq, unique int
					var idxName string
					if err := idxRows.Scan(&seq, &idxName, &unique); err != nil {
						continue
					}
					if unique == 0 {
						continue
					}
					infoRows, _ := conn.Query(fmt.Sprintf(`PRAGMA index_info("%s")`, idxName))
					for infoRows.Next() {
						var idxSeq, cid int
						var colName string
						infoRows.Scan(&idxSeq, &cid, &colName)
						for i := range ti.Columns {
							if ti.Columns[i].Name == colName {
								ti.Columns[i].IsUnique = true
								break
							}
						}
					}
					infoRows.Close()
				}
				idxRows.Close()
			}

			// Foreign keys
			fkRows, err := conn.Query(fmt.Sprintf(`PRAGMA foreign_key_list("%s")`, tbl))
			if err == nil {
				for fkRows.Next() {
					var (
						id, seq                   int
						refTable                  string
						fromCol, toCol            string
						onUpdate, onDelete, match string
					)
					if err := fkRows.Scan(&id, &seq, &refTable, &fromCol, &toCol, &onUpdate, &onDelete, &match); err != nil {
						continue
					}
					for i := range ti.Columns {
						if ti.Columns[i].Name == fromCol {
							ti.Columns[i].ForeignTable = sql.NullString{String: refTable, Valid: true}
							ti.Columns[i].ForeignColumn = sql.NullString{String: toCol, Valid: true}
							break
						}
					}
				}
				fkRows.Close()
			}

			// Row count
			var cnt int
			if err := conn.QueryRow(fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, tbl)).Scan(&cnt); err != nil {
				log.Printf("row count %s: %v", tbl, err)
			}
			ti.RowCount = cnt

			schema[tbl] = ti
		}

		return schema, nil
	}

	// PostgreSQL & MySQL branch
	var colQuery, pkQuery, uqQuery, fkQuery, rowCountFormat string

	switch provider {
	case "postgres", "postgresql":
		colQuery = PostgresQueries.Columns
		pkQuery = PostgresQueries.PrimaryKeys
		uqQuery = PostgresQueries.UniqueKeys
		fkQuery = PostgresQueries.ForeignKeys
		rowCountFormat = PostgresQueries.RowCount

	case "mysql":
		colQuery = MySQLQueries.Columns
		pkQuery = MySQLQueries.PrimaryKeys
		uqQuery = MySQLQueries.UniqueKeys
		fkQuery = MySQLQueries.ForeignKeys
		rowCountFormat = MySQLQueries.RowCount

	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	// get columns info
	rows, err := conn.Query(colQuery)
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
	pkRows, err := conn.Query(pkQuery)
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
	uqRows, err := conn.Query(uqQuery)
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
	fkRows, err := conn.Query(fkQuery)
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
		var query string

		switch provider {
		case "postgres", "postgresql":
			query = fmt.Sprintf(rowCountFormat, pq.QuoteIdentifier(tbl))
		case "mysql":
			query = fmt.Sprintf(rowCountFormat, tbl)
		}

		if err := conn.QueryRow(query).Scan(&cnt); err != nil {
			log.Printf("row count %s: %v", tbl, err)
		}
		ti.RowCount = cnt
		schema[tbl] = ti
	}

	return schema, nil
}
