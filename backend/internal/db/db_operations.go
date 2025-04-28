package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

var GetDatabases = getDatabases

// GetDatabases returns non-template Postgres DB names.
func getDatabases(provider string, conn *sql.DB) ([]string, error) {
	var rows *sql.Rows
	var err error

	switch provider {
	case "postgresql", "": // Default to PostgreSQL
		rows, err = conn.Query(`SELECT datname FROM pg_database WHERE datistemplate = false;`)
	case "mysql":
		rows, err = conn.Query(`SHOW DATABASES;`)
	default:
		return nil, fmt.Errorf("unsupported database provider: %s", provider)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, nil
}

// CreateDatabase issues CREATE DATABASE <name>.
func CreateDatabase(conn *sql.DB, name, provider string) error {
	var stmt string

	switch provider {
	case "postgresql", "":
		stmt = fmt.Sprintf("CREATE DATABASE %s;", pq.QuoteIdentifier(name))
	case "mysql":
		stmt = fmt.Sprintf("CREATE DATABASE `%s`;", name)
	default:
		return fmt.Errorf("unsupported database provider: %s", provider)
	}

	_, err := conn.Exec(stmt)
	return err
}

// DeleteDatabase issues DROP DATABASE <name>.
func DeleteDatabase(conn *sql.DB, name, provider string) error {
	var stmt string

	switch provider {
	case "postgresql", "":
		stmt = fmt.Sprintf("DROP DATABASE %s;", pq.QuoteIdentifier(name))
	case "mysql":
		stmt = fmt.Sprintf("DROP DATABASE `%s`;", name)
	default:
		return fmt.Errorf("unsupported database provider: %s", provider)
	}

	_, err := conn.Exec(stmt)
	if err != nil {
		if strings.Contains(err.Error(), "being accessed by other users") || strings.Contains(err.Error(), "Cannot drop the database") {
			return fmt.Errorf("database %s is in use", name)
		}
	}
	return err
}
