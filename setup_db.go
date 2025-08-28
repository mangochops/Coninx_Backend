
package setup

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func setup() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load DB connection string from environment variable
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable not set")
	}

	// Connect to PostgreSQL
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	// Read the schema.sql file
	sqlContent, err := ioutil.ReadFile("schema.sql")
	if err != nil {
		log.Fatalf("Unable to read schema.sql: %v\n", err)
	}

	// Execute the SQL script
	_, err = conn.Exec(context.Background(), string(sqlContent))
	if err != nil {
		log.Fatalf("Unable to execute schema: %v\n", err)
	}

	fmt.Println("Database schema created successfully!")
}
