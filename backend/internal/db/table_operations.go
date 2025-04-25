package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func ExecuteModification(ctx context.Context, conn *sql.DB, sqlQuery string) (int64, error) {
	normalized := strings.ToUpper(strings.TrimSpace(sqlQuery))

	if strings.HasPrefix(normalized, "DROP DATABASE") {
		return 0, fmt.Errorf("DROP DATABASE is not allowed")
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	done := make(chan struct{})
	var affected int64

	go func() {
		defer close(done)
		res, execErr := tx.ExecContext(ctx, sqlQuery)
		if execErr != nil {
			err = execErr
			return
		}

		affected, err = res.RowsAffected()
		if err != nil {
			return
		}

		err = tx.Commit()
	}()

	select {
	case <-ctx.Done():
		tx.Rollback()
		return 0, fmt.Errorf("request was cancelled: %w", ctx.Err())
	case <-done:
		if err != nil {
			tx.Rollback()
			return 0, err
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
