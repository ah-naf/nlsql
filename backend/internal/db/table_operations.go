package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func ExecuteModification(ctx context.Context, conn *sql.DB, sqlQuery string) (int64, error) {
	normalized := strings.ToUpper(strings.TrimSpace(sqlQuery))

	if strings.HasPrefix(normalized, "DROP DATABASE") {
		return 0, fmt.Errorf("DROP DATABASE is not allowed")
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := conn.BeginTx(ctxWithTimeout, nil)
	if err != nil {
		return 0, err
	}

	done := make(chan struct{})
	var affected int64
	var execErr error

	go func() {
		defer close(done)
		res, err := tx.ExecContext(ctx, sqlQuery)
		if err != nil {
			execErr = err
			return
		}

		affected, err = res.RowsAffected()
		if err != nil {
			execErr = err
			return
		}

		execErr = tx.Commit()
	}()

	select {
	case <-ctxWithTimeout.Done():
		tx.Rollback()
		return 0, fmt.Errorf("execution timeout or cancelled: %w", ctxWithTimeout.Err())
	case <-done:
		if execErr != nil {
			tx.Rollback()
			return 0, execErr
		}
		return affected, nil
	}
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
