// db/connection.go

package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"nlsql/internal/models"
)

var OpenConnection = openConnection

// OpenConnection opens a *sql.DB to the given database (defaults to "postgres" if DBName=="").
func openConnection(conf models.DBRequest) (*sql.DB, error) {
	if conf.ConnectionString != "" {
		driver := conf.Provider
		if driver == "" {
			driver = getDriverNameFromConnectionString(conf.ConnectionString)
		}
		// normalize
		if driver == "postgresql" {
			driver = "postgres"
		}
		return sql.Open(driver, conf.ConnectionString)
	}

	if conf.Host == "" || conf.User == "" || conf.Pass == "" {
		return nil, fmt.Errorf("required connection parameters are missing")
	}

	if conf.Port == "" {
		conf.Port = getPort(conf.Provider)
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

var OpenAdminConnection = openAdminConnection

// OpenAdminConnection connects always to the "postgres" DB for create/delete operations.
func openAdminConnection(conf models.DBRequest) (*sql.DB, error) {
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

func getPort(provider string) string {
	switch provider {
	case "postgres", "postgresql":
		return "5432"
	case "mysql":
		return "3306"
	default:
		// default to postgres
		return "5432"
	}
}
