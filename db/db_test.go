package db

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/nathanbizkit/article-management/env"
	"github.com/nathanbizkit/article-management/test"
	"github.com/stretchr/testify/assert"
)

func TestUnit_DB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests.")
	}

	lct := test.NewLocalTestContainer(t)

	t.Run("New", func(t *testing.T) {
		tests := []struct {
			title    string
			environ  *env.ENV
			hasError bool
		}{
			{
				"new db connection: success",
				&env.ENV{
					DBUser: lct.Environ().DBUser,
					DBPass: lct.Environ().DBPass,
					DBHost: lct.Environ().DBHost,
					DBPort: lct.Environ().DBPort,
					DBName: lct.Environ().DBName,
				},
				false,
			},
			{
				"new db connection: unknown db",
				&env.ENV{
					DBUser: "unknown_user",
					DBPass: "unknown_password",
					DBHost: "unknown_host",
					DBPort: "5432",
					DBName: "unknown_db",
				},
				true,
			},
		}

		for _, tt := range tests {
			actualDB, err := New(tt.environ)

			if tt.hasError {
				assert.Error(t, err, tt.title)
				assert.Empty(t, actualDB, tt.title)
			} else {
				assert.NoError(t, err, tt.title)
				assert.NotEmpty(t, actualDB, tt.title)
			}
		}
	})

	t.Run("RunInTx", func(t *testing.T) {
		tests := []struct {
			title         string
			transactionFn func(tx *sql.Tx) error
			hasError      bool
		}{
			{
				"run in tx: commit",
				func(tx *sql.Tx) error {
					_, err := tx.Query(`SELECT 1`)
					return err
				},
				false,
			},
			{
				"run in tx: rollback",
				func(tx *sql.Tx) error {
					return errors.New("transaction error")
				},
				true,
			},
		}

		for _, tt := range tests {
			err := RunInTx(lct.DB(), tt.transactionFn)

			if tt.hasError {
				assert.Error(t, err, tt.title)
			} else {
				assert.NoError(t, err, tt.title)
			}
		}
	})
}
