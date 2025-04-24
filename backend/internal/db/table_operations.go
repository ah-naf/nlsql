package db

import (
	"database/sql"
	"fmt"
	"strings"
)

func ExecuteModification(conn *sql.DB, sqlQuery string) (int64, error) {
	normalized := strings.ToUpper(strings.TrimSpace(sqlQuery))

	if strings.HasPrefix(normalized, "DROP DATABASE") {
		return 0, fmt.Errorf("DROP DATABASE is not allowed")
	}

	tx, err := conn.Begin()
	if err != nil {
		return 0, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	res, err := tx.Exec(sqlQuery)
	if err != nil {
		return 0, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, nil
	}

	return affected, nil
}

func ExecuteQuery(conn *sql.DB, sqlQuery string) ([]map[string]interface{}, error) {
	rows, err := conn.Query(sqlQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	results := []map[string]interface{}{}

	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range ptrs {
			ptrs[i] = &vals[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			continue
		}

		row := map[string]interface{}{}
		for i, n := range cols {
			row[n] = vals[i]
		}
		results = append(results, row)
	}

	return results, nil
}
