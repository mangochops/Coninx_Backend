package Admin

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var dbPool *pgxpool.Pool // shared across the Admin package

// InitDBPool initializes the DB connection pool
func InitDBPool(connString string) error {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return err
	}

	// ✅ Fix for pgx v5: disable prepared statements globally
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return err
	}

	dbPool = pool
	log.Println("[DB] Connection pool initialized")
	return nil
}

// GetDB gives access to the dbPool
func GetDB() *pgxpool.Pool {
	return dbPool
}


