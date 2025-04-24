package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"nlsql/internal/models"
)

// OpenConnection opens a *sql.DB to the given database (defaults to "postgres" if DBName=="").
func OpenConnection(conf models.DBRequest) (*sql.DB, error) {
	if conf.DBName == "" {
		conf.DBName = "postgres"
	}
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		conf.Host, conf.Port, conf.User, conf.Pass, conf.DBName,
	)
	return sql.Open("postgres", connStr)
}

// OpenAdminConnection connects always to the "postgres" DB for create/delete operations.
func OpenAdminConnection(conf models.DBRequest) (*sql.DB, error) {
	adm := conf
	adm.DBName = "postgres"
	return OpenConnection(adm)
}
