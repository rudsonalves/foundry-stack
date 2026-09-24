package main

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/rudsonalves/foundry-stack/api/internal/bootstrap"
)

func TestOpenDatabase(t *testing.T) {
	t.Cleanup(func() {
		sqlOpen = sql.Open
		databaseDriverName = "pgx"
		databasePingTimeout = 5 * time.Second
	})

	cfg := bootstrap.DBConfig{
		URL:             "postgres://foundry_stack:secret@localhost:5432/foundry_stack",
		MaxOpenConns:    15,
		MaxIdleConns:    7,
		ConnMaxIdleTime: 2 * time.Minute,
		ConnMaxLifetime: 12 * time.Minute,
	}

	t.Run("returns wrapped error when open fails", func(t *testing.T) {
		sqlOpen = func(driverName string, dataSourceName string) (*sql.DB, error) {
			if driverName != databaseDriverName {
				t.Fatalf("driverName = %q, want %q", driverName, databaseDriverName)
			}
			if dataSourceName != cfg.URL {
				t.Fatalf("dataSourceName = %q, want %q", dataSourceName, cfg.URL)
			}
			return nil, errors.New("driver unavailable")
		}

		db, err := openDatabase(cfg)
		if err == nil {
			t.Fatal("openDatabase() expected error, got nil")
		}
		if db != nil {
			t.Fatal("openDatabase() db should be nil on error")
		}
		if got := err.Error(); !strings.Contains(got, "open database: driver unavailable") {
			t.Fatalf("error = %q, want wrapped open error", got)
		}
	})

	t.Run("closes connection when ping fails", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		if err != nil {
			t.Fatalf("sqlmock.New() unexpected error: %v", err)
		}
		defer db.Close()

		mock.ExpectPing().WillReturnError(errors.New("ping failed"))
		mock.ExpectClose()

		sqlOpen = func(driverName string, dataSourceName string) (*sql.DB, error) {
			return db, nil
		}

		gotDB, gotErr := openDatabase(cfg)
		if gotErr == nil {
			t.Fatal("openDatabase() expected ping error, got nil")
		}
		if gotDB != nil {
			t.Fatal("openDatabase() db should be nil when ping fails")
		}
		if got := gotErr.Error(); !strings.Contains(got, "ping database: ping failed") {
			t.Fatalf("error = %q, want wrapped ping error", got)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sqlmock expectations: %v", err)
		}
	})

	t.Run("returns ready connection when ping succeeds", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		if err != nil {
			t.Fatalf("sqlmock.New() unexpected error: %v", err)
		}
		defer db.Close()

		mock.ExpectPing()

		sqlOpen = func(driverName string, dataSourceName string) (*sql.DB, error) {
			return db, nil
		}

		gotDB, gotErr := openDatabase(cfg)
		if gotErr != nil {
			t.Fatalf("openDatabase() unexpected error: %v", gotErr)
		}
		if gotDB == nil {
			t.Fatal("openDatabase() returned nil db")
		}

		if stats := gotDB.Stats(); stats.MaxOpenConnections != cfg.MaxOpenConns {
			t.Fatalf(
				"MaxOpenConnections = %d, want %d",
				stats.MaxOpenConnections,
				cfg.MaxOpenConns,
			)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sqlmock expectations: %v", err)
		}
	})
}
