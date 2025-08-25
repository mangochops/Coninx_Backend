package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/mangochops/coninx_backend/Admin"
	"github.com/mangochops/coninx_backend/Driver"
	
	
)

func main() {
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
	fmt.Println("Connected to database")

	// Set up router
	r := mux.NewRouter()

	// Pass the db connection to your dispatch routes

	// Admin.auth
	Admin.RegisterAuthRoutes(r)
	// Admin.dispatch
	Admin.RegisterDispatchRoutes(r)

	// Driver.auth
	Driver.RegisterDriverRoutes(r)
	// Driver.trip
	Driver.RegisterTripRoutes(r)

	// Driver.delivery
	Driver.RegisterDeliveryRoutes(r)

	// Start server
	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
