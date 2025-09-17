package Admin

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var dbPool *pgxpool.Pool // shared across the Admin package

// InitDBPool initializes the DB connection pool
func InitDBPool(connString string) error {
	pool, err := pgxpool.New(context.Background(), connString)
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
