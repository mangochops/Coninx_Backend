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
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable not set")
	}

	// Connect to PostgreSQL (Neon)
conn, err := pgx.Connect(context.Background(), dbURL)
if err != nil {
	log.Fatalf("Unable to connect to database: %v\n", err)
}
fmt.Println("Connected to database")

// Ensure connection is closed when main exits
defer func() {
	if err := conn.Close(context.Background()); err != nil {
		log.Printf("Error closing database connection: %v\n", err)
	}
}()

// Set search path to public schema (Neon)
_, err = conn.Exec(context.Background(), "SET search_path TO public")
if err != nil {
	log.Fatalf("Failed to set search path: %v\n", err)
}

// Initialize DB connections for packages
Admin.InitDB(conn)
Driver.InitDB(conn) // Ensure Driver package has InitDB function


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

	// --- Admin routes ---
	Admin.RegisterAuthRoutes(adminRouter)      // /admin/register, /admin/login
	Admin.RegisterDispatchRoutes(adminRouter)  // /admin/dispatches
	Admin.RegisterVehicleRoutes(adminRouter)   // /admin/vehicles
	Admin.RegisterTripRoutes(adminRouter)      // /admin/trips

	// --- Driver routes ---
	Driver.RegisterDriverRoutes(driverRouter)      // /driver/drivers
	Driver.RegisterTripRoutes(driverRouter)        // /driver/trips
	Driver.RegisterDeliveryRoutes(driverRouter)    // /driver/deliveries

	// Debug: Print all registered routes
	err = r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		fmt.Printf("Route registered: %s %v\n", path, methods)
		return nil
	})
	if err != nil {
		log.Println("Error walking routes:", err)
	}

	// Start server
	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}


