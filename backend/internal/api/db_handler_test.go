package api_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"nlsql/internal/api"
	"nlsql/internal/db"
	"nlsql/internal/models"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGetDatabases_OK(t *testing.T) {
	origOpen := db.OpenConnection
	origList := db.GetDatabases

	t.Cleanup(func() {
		db.OpenConnection = origOpen
		db.GetDatabases = origList
	})

	db.OpenConnection = func(req models.DBRequest, c *gin.Context) (*sql.DB, error) {
		mockDB, _, _ := sqlmock.New() // mock driver automatically registered
		return mockDB, nil            // Close() is perfectly safe
	}
	db.GetDatabases = func(provider string, _ *sql.DB) ([]string, error) {
		return []string{"db1", "db2"}, nil
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/databases?host=h&port=p&user=u&pass=x&provider=postgres", nil)

	api.GetDatabases(c)

	require.Equal(t, http.StatusOK, w.Code)

	var got models.DatabaseListResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Equal(t, []string{"db1", "db2"}, got.Databases)
}

func TestGetDatabases_MissingParams(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/databases", nil)

	api.GetDatabases(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateDB_OK(t *testing.T) {
	origOpenAdmin := db.OpenAdminConnection
	origCreateDB := db.CreateDatabase
	t.Cleanup(func() {
		db.OpenAdminConnection = origOpenAdmin
		db.CreateDatabase = origCreateDB
	})

	/* ---- stub DB layer ---- */
	db.OpenAdminConnection = func(req models.DBRequest, c *gin.Context) (*sql.DB, error) {
		mockDB, _, _ := sqlmock.New()
		return mockDB, nil
	}
	db.CreateDatabase = func(_ *sql.DB, name, _ string) error {
		return nil
	}

	/* ---- craft request ---- */
	body, _ := json.Marshal(models.DBRequest{
		Host: "h", Port: "p", User: "u", Pass: "x",
		Provider: "postgres", DBName: "reporting",
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	/* ---- call handler ---- */
	api.CreateDB(c)

	/* ---- assertions ---- */
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "Database created", resp["message"])
	require.Equal(t, "reporting", resp["dbname"])
}

func TestCreateDB_CreateDatabaseError(t *testing.T) {
	origOpen := db.OpenAdminConnection
	origCD := db.CreateDatabase
	t.Cleanup(func() {
		db.OpenAdminConnection = origOpen
		db.CreateDatabase = origCD
	})

	mockDB, _, _ := sqlmock.New()
	db.OpenAdminConnection = func(models.DBRequest, *gin.Context) (*sql.DB, error) { return mockDB, nil }
	db.CreateDatabase = func(*sql.DB, string, string) error { return errors.New("syntax error") }

	body := `{"dbname":"bad","provider":"postgres"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	api.CreateDB(c)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.Contains(t, w.Body.String(), "syntax error")
}
