package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"

	"nlsql/internal/models"
)

// OpenConnection opens a *sql.DB to the given database (defaults to "postgres" if DBName=="").
func OpenConnection(conf models.DBRequest) (*sql.DB, error) {
	if conf.ConnectionString != "" {
		return sql.Open(getDriverNameFromConnectionString(conf.ConnectionString), conf.ConnectionString)
	}

	if conf.Host == "" || conf.Port == "" || conf.User == "" || conf.Pass == "" {
		return nil, fmt.Errorf("required connection parameters are missing")
	}

	if conf.DBName == "" {
		switch conf.Provider {
		case "postgresql", "":
			conf.DBName = "postgres"
		case "mysql":
			conf.DBName = "mysql"
		case "mssql":
			conf.DBName = "master"
		default:
			conf.DBName = "postgres" // Default to postgres
		}
	}

	if conf.SSLMode == "" {
		conf.SSLMode = "disable"
	}

	switch conf.Provider {
	case "postgresql", "": // Default to PostgreSQL if not specified
		connStr := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			conf.Host, conf.Port, conf.User, conf.Pass, conf.DBName, conf.SSLMode,
		)
		return sql.Open("postgres", connStr)

	case "mysql":
		// Format: username:password@tcp(host:port)/dbname
		connStr := fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s",
			conf.User, conf.Pass, conf.Host, conf.Port, conf.DBName,
		)
		return sql.Open("mysql", connStr)

	default:
		return nil, fmt.Errorf("unsupported database provider: %s", conf.Provider)
	}
}

func getDriverNameFromConnectionString(connStr string) string {
	if strings.HasPrefix(connStr, "postgresql://") || strings.HasPrefix(connStr, "postgres://") {
		return "postgres"
	} else if strings.HasPrefix(connStr, "mysql://") {
		return "mysql"
	} else if strings.HasPrefix(connStr, "sqlserver://") {
		return "mssql"
	} else if strings.HasPrefix(connStr, "mongodb://") {
		return "mongodb"
	} else {
		return "postgres"
	}
}

// OpenAdminConnection connects always to the "postgres" DB for create/delete operations.
func OpenAdminConnection(conf models.DBRequest) (*sql.DB, error) {
	adminConf := conf

	switch conf.Provider {
	case "postgresql", "":
		adminConf.DBName = "postgres" // System database for PostgreSQL
	case "mysql":
		adminConf.DBName = "mysql" // System database for MySQL
	case "mssql":
		adminConf.DBName = "master" // System database for SQL Server
	}

	return OpenConnection(adminConf)
}
