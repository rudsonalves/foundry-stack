package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rudsonalves/foundry-stack/api/internal/bootstrap"
)

var (
	sqlOpen             = sql.Open
	databaseDriverName  = "pgx"
	databasePingTimeout = 5 * time.Second
)

func openDatabase(dbConfig bootstrap.DBConfig) (*sql.DB, error) {
	db, err := sqlOpen(databaseDriverName, dbConfig.URL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(dbConfig.MaxOpenConns)
	db.SetMaxIdleConns(dbConfig.MaxIdleConns)
	db.SetConnMaxIdleTime(dbConfig.ConnMaxIdleTime)
	db.SetConnMaxLifetime(dbConfig.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		databasePingTimeout,
	)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}
