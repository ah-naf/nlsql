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
