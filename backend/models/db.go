package models

import "database/sql"

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
