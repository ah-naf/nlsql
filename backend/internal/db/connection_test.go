package db_test

import (
	"testing"

	"nlsql/internal/db"
	"nlsql/internal/models"

	"github.com/stretchr/testify/require"
)

func TestOpenConnection(t *testing.T) {
	type tc struct {
		name    string
		req     models.DBRequest
		wantErr bool
	}
	cases := []tc{
		{
			name: "full params – postgres",
			req:  models.DBRequest{Provider: "postgresql", Host: "h", Port: "5432", User: "u", Pass: "x", DBName: "app"},
		},
		{
			name: "full params – mysql",
			req:  models.DBRequest{Provider: "mysql", Host: "h", Port: "3306", User: "u", Pass: "x", DBName: "app"},
		},
		{
			name: "connection string infers driver – postgres",
			req:  models.DBRequest{ConnectionString: "postgres://u:x@localhost:5432/app?sslmode=disable"},
		},
		{
			name: "connection string infers driver – mysql",
			req:  models.DBRequest{ConnectionString: "mysql://u:x@tcp(localhost:3306)/app"},
		},
		{
			name:    "unsupported provider",
			req:     models.DBRequest{Provider: "shifat", Host: "h", Port: "1521", User: "scott", Pass: "tiger"},
			wantErr: true,
		},
		{
			name:    "missing required params",
			req:     models.DBRequest{Provider: "postgresql", Host: "h", Port: "5432", Pass: "x"}, // no user
			wantErr: true,
		},
	}

	for _, tt := range cases {
		tt := tt // capture
		t.Run(tt.name, func(t *testing.T) {
			dbh, err := db.OpenConnection(tt.req, nil)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, dbh)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, dbh)
			_ = dbh.Close()
		})
	}
}

func TestOpenAdminConnection(t *testing.T) {
	type tc struct {
		name    string
		req     models.DBRequest
		wantErr bool
	}
	cases := []tc{
		{
			name:    "postgres provider – overrides dbname",
			req:     models.DBRequest{Provider: "postgresql", Host: "h", Port: "5432", User: "u", Pass: "x", DBName: "custom"},
			wantErr: false,
		},
		{
			name:    "mysql provider – overrides dbname",
			req:     models.DBRequest{Provider: "mysql", Host: "h", Port: "3306", User: "u", Pass: "x", DBName: "custom"},
			wantErr: false,
		},
		{
			name:    "unsupported provider",
			req:     models.DBRequest{Provider: "oracle", Host: "h", Port: "1521", User: "u", Pass: "x"},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			dbh, err := db.OpenAdminConnection(tt.req, nil)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, dbh)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, dbh)

			_ = dbh.Close()
		})
	}
}
