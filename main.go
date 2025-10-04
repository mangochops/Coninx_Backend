package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/mangochops/coninx_backend/Admin"
	"github.com/mangochops/coninx_backend/Driver"
	"github.com/rs/cors"
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

	// Connect to PostgreSQL using pgxpool
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}
	fmt.Println("Connected to database")

	defer pool.Close()

	// Set search path to public schema
	_, err = pool.Exec(context.Background(), "SET search_path TO public")
	if err != nil {
		log.Fatalf("Failed to set search path: %v\n", err)
	}

	// Initialize DB connections for packages
	Admin.InitDBPool(dbURL)
	Driver.InitDB(pool)

	// Test the connection
	var result int
	err = pool.QueryRow(context.Background(), "SELECT 1").Scan(&result)
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

	// ✅ Enable CORS properly
	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"https://conninx-dashboard.vercel.app", // production
			"https://*.vercel.app",                 // Vercel preview deployments
			"http://localhost:3000",                // local dev
			"http://127.0.0.1:3000",
			"http://127.0.0.1:8000",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            true, // shows preflight logs in terminal
	})

	// ✅ Use Render's dynamic PORT instead of hardcoding 5000
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	fmt.Printf("Server running on :%s\n", port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, c.Handler(router)))
}
