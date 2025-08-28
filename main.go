package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/mangochops/coninx_backend/Admin"
	"github.com/mangochops/coninx_backend/Driver"
)

func main() {
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
	fmt.Println("Connected to database")

	// Pass the db connection to Admin package
	Admin.InitDB(conn)

	// Test the connection with a simple query
	var result int
	err = conn.QueryRow(context.Background(), "SELECT 1").Scan(&result)
	if err != nil {
		log.Printf("Failed to execute test query: %v\n", err)
	} else {
		fmt.Printf("Database test query successful! Result: %d\n", result)
	}

	// Set up router
	r := mux.NewRouter()

	// --- Namespaced routers ---
	adminRouter := r.PathPrefix("/admin").Subrouter()
	driverRouter := r.PathPrefix("/driver").Subrouter()

	// Admin routes
	Admin.RegisterAuthRoutes(adminRouter)
	Admin.RegisterDispatchRoutes(adminRouter)
	Admin.RegisterVehicleRoutes(adminRouter)
	Admin.RegisterTripRoutes(adminRouter)

	// Driver routes
	Driver.RegisterDriverRoutes(driverRouter)
	Driver.RegisterTripRoutes(driverRouter)
	Driver.RegisterDeliveryRoutes(driverRouter)

	// Debug: Print registered routes
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		fmt.Printf("Route registered: %s %v\n", path, methods)
		return nil
	})

	// Start server
	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

