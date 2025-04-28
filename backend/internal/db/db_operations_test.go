package db_test

import (
	"errors"
	"nlsql/internal/db"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestCreateDatabase(t *testing.T) {
	type tc struct {
		name      string
		provider  string
		stmtRegex string
		wantErr   bool
	}
	cases := []tc{
		{
			name:      "postgres – success",
			provider:  "postgresql",
			stmtRegex: `CREATE DATABASE "app";`,
		},
		{
			name:      "mysql – success",
			provider:  "mysql",
			stmtRegex: "CREATE DATABASE `app`;",
		},
		{
			name:     "unsupported provider",
			provider: "oracle",
			wantErr:  true,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			dbh, mock, _ := sqlmock.New()
			defer dbh.Close()

			if !tt.wantErr {
				mock.ExpectExec(regexp.QuoteMeta(tt.stmtRegex)).
					WillReturnResult(sqlmock.NewResult(0, 1))
			}

			err := db.CreateDatabase(dbh, "app", tt.provider)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}

func TestDeleteDatabase(t *testing.T) {
	type tc struct {
		name         string
		provider     string
		stmtRegex    string
		mockErr      error
		expectInUse  bool
		expectDrvErr bool
		expectUnsupp bool
	}
	cases := []tc{
		{
			name:      "postgres – success",
			provider:  "postgresql",
			stmtRegex: `DROP DATABASE "app";`,
		},
		{
			name:      "mysql – success",
			provider:  "mysql",
			stmtRegex: "DROP DATABASE `app`;",
		},
		{
			name:        "postgres – in use error mapped",
			provider:    "postgresql",
			stmtRegex:   `DROP DATABASE "app";`,
			mockErr:     errors.New("database is being accessed by other users"),
			expectInUse: true,
		},
		{
			name:         "unsupported provider",
			provider:     "oracle",
			expectUnsupp: true,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			dbh, mock, _ := sqlmock.New()
			defer dbh.Close()

			if tt.expectUnsupp {
				err := db.DeleteDatabase(dbh, "app", tt.provider)
				require.Error(t, err)
				require.Contains(t, err.Error(), "unsupported")
				return
			}

			exec := mock.ExpectExec(regexp.QuoteMeta(tt.stmtRegex))
			if tt.mockErr != nil {
				exec.WillReturnError(tt.mockErr)
			} else {
				exec.WillReturnResult(sqlmock.NewResult(0, 1))
			}

			err := db.DeleteDatabase(dbh, "app", tt.provider)

			if tt.expectInUse {
				require.Error(t, err)
				require.Contains(t, err.Error(), "in use")
			} else {
				require.NoError(t, err)
				require.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}
