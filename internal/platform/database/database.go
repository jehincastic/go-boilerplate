package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jehincastic/go-boilerplate/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type txKey struct{}

// Open connects to PostgreSQL, applies pool limits, and pings the database.
func Open(ctx context.Context, cfg config.Database, env string) (*gorm.DB, error) {
	logLevel := gormlogger.Warn
	if env == "development" {
		logLevel = gormlogger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
		Logger:                 gormlogger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("database pool: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxIdleTime(cfg.MaxIdleTime)
	sqlDB.SetConnMaxLifetime(time.Hour)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		if closeErr := sqlDB.Close(); closeErr != nil {
			return nil, fmt.Errorf("ping database: %w (close: %v)", err, closeErr)
		}
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

// Ping checks that the pool can reach PostgreSQL.
func Ping(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// WithTx stores a transaction on ctx so repositories share one unit of work.
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// FromContext returns the active transaction, or db when there is none.
func FromContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok && tx != nil {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}

// WithinTransaction runs fn inside a database transaction.
func WithinTransaction(ctx context.Context, db *gorm.DB, fn func(ctx context.Context) error) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(WithTx(ctx, tx))
	})
}

// WithTimeout bounds a repository call without dropping a transaction stored on ctx.
func WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 3*time.Second)
}
