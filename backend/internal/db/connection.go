// db/connection.go

package db

import (
	"database/sql"
	"fmt"
	"strings"

	"nlsql/internal/models"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

var OpenConnection = openConnection

// OpenConnection opens a *sql.DB to the given database (defaults to "postgres" if DBName=="").
func openConnection(conf models.DBRequest) (*sql.DB, error) {
	// Handle if the user want demo database
	if conf.Provider == "demo" {
		conn, err := sql.Open("sqlite", ":memory:")
		if err != nil {
			return nil, err
		}

		if err := initDemoSchema(conn); err != nil {
			conn.Close()
			return nil, err
		}
		return conn, nil
	}

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

func initDemoSchema(conn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE users (
            id         INTEGER PRIMARY KEY,
            name       TEXT    NOT NULL,
            email      TEXT    NOT NULL UNIQUE,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );`,

		`CREATE TABLE products (
            id         INTEGER PRIMARY KEY,
            name       TEXT    NOT NULL UNIQUE,
            price      REAL    NOT NULL CHECK(price >= 0),
            stock      INTEGER NOT NULL DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );`,

		`CREATE TABLE categories (
            id   INTEGER PRIMARY KEY,
            name TEXT NOT NULL UNIQUE
        );`,

		`CREATE TABLE product_categories (
            product_id  INTEGER NOT NULL,
            category_id INTEGER NOT NULL,
            PRIMARY KEY (product_id, category_id),
            FOREIGN KEY (product_id)  REFERENCES products(id)   ON DELETE CASCADE,
            FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
        );`,

		`CREATE TABLE orders (
            id          INTEGER PRIMARY KEY,
            user_id     INTEGER NOT NULL,
            total       REAL    NOT NULL CHECK(total >= 0),
            status      TEXT    NOT NULL DEFAULT 'pending',
            created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (user_id) REFERENCES users(id)
        );`,

		`CREATE TABLE order_items (
            id         INTEGER PRIMARY KEY,
            order_id   INTEGER NOT NULL,
            product_id INTEGER NOT NULL,
            quantity   INTEGER NOT NULL CHECK(quantity > 0),
            unit_price REAL    NOT NULL CHECK(unit_price >= 0),
            FOREIGN KEY (order_id)   REFERENCES orders(id)   ON DELETE CASCADE,
            FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT
        );`,

		`INSERT INTO users (name, email) VALUES
            ('Alice',   'alice@example.com'),
            ('Bob',     'bob@example.com'),
            ('Charlie', 'charlie@example.com');`,

		`INSERT INTO products (name, price, stock) VALUES
            ('Gizmo',    19.99, 100),
            ('Widget',   29.95,  50),
            ('Doohickey',9.50,  200);`,

		`INSERT INTO categories (name) VALUES
            ('Gadgets'),
            ('Tools'),
            ('Accessories');`,

		`INSERT INTO product_categories (product_id, category_id) VALUES
            (1,1),(2,1),(3,3),(2,2);`,

		`INSERT INTO orders (user_id, total, status) VALUES
            (1, 49.94, 'completed'),
            (2,  9.50, 'pending');`,

		`INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES
            (1, 1, 2, 19.99),
            (1, 2, 1, 29.95),
            (2, 3, 1, 9.50);`,
	}

	for _, s := range stmts {
		if _, err := conn.Exec(s); err != nil {
			return err
		}
	}
	return nil
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
