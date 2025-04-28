package db

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestExecuteModification_Success(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()

	const q = "UPDATE users SET active = false WHERE last_login < NOW() - INTERVAL '30 days'"

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(q)).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	affected, err := ExecuteModification(context.Background(), mockDB, q)

	require.NoError(t, err)
	require.Equal(t, int64(2), affected)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExecuteModification_DropDatabaseBlocked(t *testing.T) {
	mockDB, _, _ := sqlmock.New()
	defer mockDB.Close()

	affected, err := ExecuteModification(context.Background(), mockDB, "DROP DATABASE prod")

	require.Error(t, err)
	require.Contains(t, err.Error(), "not allowed")
	require.Equal(t, int64(0), affected)
}

func TestExecuteModification_ExecError(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()

	const q = "DELETE FROM users WHERE id = 1"

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(q)).
		WillReturnError(errors.New("syntax error"))
	mock.ExpectRollback()

	affected, err := ExecuteModification(context.Background(), mockDB, q)

	require.Error(t, err)
	require.Equal(t, int64(0), affected)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExecuteQuery_Success(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()

	const q = "SELECT id, name FROM users"

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "Alice").
		AddRow(2, "Bob")

	mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnRows(rows)

	got, err := ExecuteQuery(mockDB, q)

	require.NoError(t, err)
	require.Len(t, got, 2)

	require.Equal(t, map[string]interface{}{"id": int64(1), "name": "Alice"}, got[0])
	require.Equal(t, map[string]interface{}{"id": int64(2), "name": "Bob"}, got[1])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExecuteQuery_QueryError(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()

	const q = "SELECT * FROM does_not_exist"

	mock.ExpectQuery(regexp.QuoteMeta(q)).
		WillReturnError(errors.New("relation does_not_exist does not exist"))

	res, err := ExecuteQuery(mockDB, q)

	require.Error(t, err)
	require.Nil(t, res)
	require.NoError(t, mock.ExpectationsWereMet())
}
