// controller/schema.go
package controller

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// ColumnInfo, TableInfo types here exactly as you had them
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

// GetSchema handles both brief (names+counts) and full schema dumps
func GetSchema(c *gin.Context) {
	req := DBRequest{
		Host:   c.Query("host"),
		Port:   c.Query("port"),
		User:   c.Query("user"),
		Pass:   c.Query("pass"),
		DBName: c.Query("dbname"),
	}

	if req.Host == "" || req.Port == "" || req.User == "" || req.Pass == "" || req.DBName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required query parameters"})
		return
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		req.Host, req.Port, req.User, req.Pass, req.DBName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Connection error: " + err.Error()})
		return
	}
	defer db.Close()

	// If ?brief=true, just return [{ name, row_count }, ...]
	if c.Query("brief") == "true" {
		rows, err := db.Query(`
			SELECT table_name
			FROM information_schema.tables
			WHERE table_schema='public'
			ORDER BY table_name
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		type briefT struct {
			Name     string `json:"name"`
			RowCount int    `json:"row_count"`
		}
		var list []briefT

		for rows.Next() {
			var tbl string
			if err := rows.Scan(&tbl); err != nil {
				log.Printf("scan name: %v", err)
				continue
			}
			var cnt int
			// note: for large tables you might switch to an estimate
			if err := db.QueryRow(
				fmt.Sprintf("SELECT COUNT(*) FROM %s", pq.QuoteIdentifier(tbl)),
			).Scan(&cnt); err != nil {
				log.Printf("count %s: %v", tbl, err)
				continue
			}
			list = append(list, briefT{tbl, cnt})
		}

		c.JSON(http.StatusOK, gin.H{"tables": list})
		return
	}

	// Otherwise return the full schema map
	full, err := loadFullSchema(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"schema": full})
}

// GetTableSchema returns detailed columns for a single table
func GetTableSchema(c *gin.Context) {
	req := DBRequest{
		Host:   c.Query("host"),
		Port:   c.Query("port"),
		User:   c.Query("user"),
		Pass:   c.Query("pass"),
		DBName: c.Query("dbname"),
	}

	if req.Host == "" || req.Port == "" || req.User == "" || req.Pass == "" || req.DBName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required query parameters"})
		return
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		req.Host, req.Port, req.User, req.Pass, req.DBName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Connection error: " + err.Error()})
		return
	}
	defer db.Close()

	table := c.Param("tableName")
	full, err := loadFullSchema(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	info, ok := full[table]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "table not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"table": info})
}

// controller/schema.go
// Method to fix the duplication issue by consolidating constraints

func loadFullSchema(db *sql.DB) (map[string]TableInfo, error) {
	// First query to get basic column info and table descriptions
	const colQuery = `
	SELECT
	  c.table_name,
	  c.column_name,
	  c.data_type,
	  (c.is_nullable = 'YES') AS is_nullable,
	  c.column_default,
	  pgd.description AS column_description,
	  tbl_pgd.description AS table_description
	FROM information_schema.columns AS c
	LEFT JOIN pg_catalog.pg_stat_all_tables AS st
	  ON c.table_name = st.relname
	LEFT JOIN pg_catalog.pg_description AS pgd
	  ON pgd.objoid = st.relid
	 AND pgd.objsubid = c.ordinal_position
	LEFT JOIN pg_catalog.pg_description AS tbl_pgd
	  ON tbl_pgd.objoid = st.relid
	 AND tbl_pgd.objsubid = 0
	WHERE c.table_schema = 'public'
	ORDER BY c.table_name, c.ordinal_position;
	`

	rows, err := db.Query(colQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schema := map[string]TableInfo{}

	// Process basic column info first
	for rows.Next() {
		var (
			tbl, col, dt     string
			nullStrDefault   sql.NullString
			colDesc, tblDesc sql.NullString
			isNullable       bool
		)
		if err := rows.Scan(
			&tbl, &col, &dt, &isNullable, &nullStrDefault,
			&colDesc, &tblDesc,
		); err != nil {
			return nil, err
		}

		cinfo := ColumnInfo{
			Name:          col,
			DataType:      dt,
			IsNullable:    isNullable,
			DefaultValue:  nullStrDefault,
			IsPrimaryKey:  false, // Will set these constraints in separate queries
			IsUnique:      false,
			ForeignTable:  sql.NullString{},
			ForeignColumn: sql.NullString{},
			Description:   colDesc,
		}

		ti := schema[tbl]
		if ti.Name == "" {
			ti = TableInfo{
				Name:        tbl,
				Columns:     []ColumnInfo{},
				Description: tblDesc,
			}
		}
		ti.Columns = append(ti.Columns, cinfo)
		schema[tbl] = ti
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Now add primary key constraints
	pkQuery := `
	SELECT tc.table_name, kcu.column_name
	FROM information_schema.table_constraints AS tc
	JOIN information_schema.key_column_usage AS kcu
	  ON tc.constraint_schema = kcu.constraint_schema
	 AND tc.constraint_name = kcu.constraint_name
	WHERE tc.constraint_type = 'PRIMARY KEY'
	  AND tc.table_schema = 'public'
	`

	pkRows, err := db.Query(pkQuery)
	if err != nil {
		return nil, err
	}
	defer pkRows.Close()

	for pkRows.Next() {
		var tbl, col string
		if err := pkRows.Scan(&tbl, &col); err != nil {
			return nil, err
		}

		// Update the column in our schema
		tableInfo, ok := schema[tbl]
		if !ok {
			continue
		}

		for i, colInfo := range tableInfo.Columns {
			if colInfo.Name == col {
				tableInfo.Columns[i].IsPrimaryKey = true
				schema[tbl] = tableInfo
				break
			}
		}
	}

	// Add unique constraints
	uniqueQuery := `
	SELECT tc.table_name, kcu.column_name
	FROM information_schema.table_constraints AS tc
	JOIN information_schema.key_column_usage AS kcu
	  ON tc.constraint_schema = kcu.constraint_schema
	 AND tc.constraint_name = kcu.constraint_name
	WHERE tc.constraint_type = 'UNIQUE'
	  AND tc.table_schema = 'public'
	`

	uniqueRows, err := db.Query(uniqueQuery)
	if err != nil {
		return nil, err
	}
	defer uniqueRows.Close()

	for uniqueRows.Next() {
		var tbl, col string
		if err := uniqueRows.Scan(&tbl, &col); err != nil {
			return nil, err
		}

		// Update the column in our schema
		tableInfo, ok := schema[tbl]
		if !ok {
			continue
		}

		for i, colInfo := range tableInfo.Columns {
			if colInfo.Name == col {
				tableInfo.Columns[i].IsUnique = true
				schema[tbl] = tableInfo
				break
			}
		}
	}

	// Finally add foreign key constraints
	fkQuery := `
	SELECT
	  kcu.table_name,
	  kcu.column_name,
	  ccu.table_name AS foreign_table_name,
	  ccu.column_name AS foreign_column_name
	FROM information_schema.table_constraints AS tc
	JOIN information_schema.key_column_usage AS kcu
	  ON tc.constraint_schema = kcu.constraint_schema
	 AND tc.constraint_name = kcu.constraint_name
	JOIN information_schema.constraint_column_usage AS ccu
	  ON tc.constraint_schema = ccu.constraint_schema
	 AND tc.constraint_name = ccu.constraint_name
	WHERE tc.constraint_type = 'FOREIGN KEY'
	  AND tc.table_schema = 'public'
	`

	fkRows, err := db.Query(fkQuery)
	if err != nil {
		return nil, err
	}
	defer fkRows.Close()

	for fkRows.Next() {
		var tbl, col, ftbl, fcol string
		if err := fkRows.Scan(&tbl, &col, &ftbl, &fcol); err != nil {
			return nil, err
		}

		// Update the column in our schema
		tableInfo, ok := schema[tbl]
		if !ok {
			continue
		}

		for i, colInfo := range tableInfo.Columns {
			if colInfo.Name == col {
				tableInfo.Columns[i].ForeignTable = sql.NullString{String: ftbl, Valid: true}
				tableInfo.Columns[i].ForeignColumn = sql.NullString{String: fcol, Valid: true}
				schema[tbl] = tableInfo
				break
			}
		}
	}

	// get row counts
	for name, ti := range schema {
		var cnt int
		if err := db.QueryRow(
			fmt.Sprintf("SELECT COUNT(*) FROM %s", pq.QuoteIdentifier(name)),
		).Scan(&cnt); err != nil {
			log.Printf("row count %s: %v", name, err)
		}
		ti.RowCount = cnt
		schema[name] = ti
	}

	return schema, nil
}
