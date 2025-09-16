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
	"github.com/rs/cors"
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

	// Connect to PostgreSQL
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	fmt.Println("Connected to database")

	defer func() {
		if err := conn.Close(context.Background()); err != nil {
			log.Printf("Error closing database connection: %v\n", err)
		}
	}()

	// Set search path to public schema
	_, err = conn.Exec(context.Background(), "SET search_path TO public")
	if err != nil {
		log.Fatalf("Failed to set search path: %v\n", err)
	}

	// Initialize DB connections for packages
	// Admin.InitDB(conn)
	Driver.InitDB(conn)

	// Test the connection
	var result int
	err = conn.QueryRow(context.Background(), "SELECT 1").Scan(&result)
	if err != nil {
		log.Printf("Failed to execute test query: %v\n", err)
	} else {
		fmt.Printf("Database test query successful! Result: %d\n", result)
	}

	// Set up router
	router := mux.NewRouter()

	// --- Namespaced routers ---
	adminRouter := router.PathPrefix("/admin").Subrouter()
	driverRouter := router.PathPrefix("/driver").Subrouter()

	// --- Admin routes ---
	Admin.RegisterAuthRoutes(adminRouter)
	Admin.RegisterDispatchRoutes(adminRouter)
	Admin.RegisterVehicleRoutes(adminRouter)
	Admin.RegisterTripRoutes(adminRouter)

	// --- Driver routes ---
	Driver.RegisterDriverRoutes(driverRouter)
	Driver.RegisterTripRoutes(driverRouter)
	Driver.RegisterDeliveryRoutes(driverRouter)

	// Debug: Print all registered routes
	err = router.Walk(func(route *mux.Route, r *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		fmt.Printf("Route registered: %s %v\n", path, methods)
		return nil
	})
	if err != nil {
		log.Println("Error walking routes:", err)
	}

	// Enable CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://conninx-dashboard.vercel.app"}, // your frontend origin
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}).Handler(router)

	// Start server
	fmt.Println("Server running on :5000")
	log.Fatal(http.ListenAndServe("0.0.0.0:5000", corsHandler))
}
