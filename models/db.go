package models

import (
    "database/sql"
    _ "github.com/lib/pq"
)

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
func GetSchema(db *sql.DB) (map[string][]string, error) {
    query := `
        SELECT table_name, column_name
        FROM information_schema.columns
        WHERE table_schema = 'public'
        ORDER BY table_name, ordinal_position;
    `
    rows, err := db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    schema := make(map[string][]string)
    for rows.Next() {
        var table, column string
        if err := rows.Scan(&table, &column); err != nil {
            return nil, err
        }
        schema[table] = append(schema[table], column)
    }
    return schema, nil
}