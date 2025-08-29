package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func SetupDB() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable not set")
	}

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

	// Execute the SQL script as-is (make sure it uses CREATE TABLE IF NOT EXISTS)
	_, err = conn.Exec(context.Background(), string(sqlContent))
	if err != nil {
		log.Fatalf("Unable to execute schema: %v\n", err)
	}

	fmt.Println("Database schema initialized successfully!")
}
