package api_test

import (
	"database/sql"
	"encoding/json"
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

	db.OpenConnection = func(req models.DBRequest) (*sql.DB, error) {
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
