package Admin

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// --- Structs ---
type Register struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// --- Global pool ---
var dbPool *pgxpool.Pool

// --- InitDBPool initializes the DB connection pool ---
func InitDBPool(connString string) error {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return err
	}
	dbPool = pool
	log.Println("[DB] Connection pool initialized")
	return nil
}

// --- SignupHandler ---
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	if dbPool == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		log.Println("[Signup] DB pool is nil")
		return
	}

	var reg Register
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		log.Println("[Signup] JSON decode error:", err)
		return
	}

	// ✅ Hash password before saving
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reg.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error securing password", http.StatusInternalServerError)
		log.Println("[Signup] Password hash error:", err)
		return
	}

	query := `
		INSERT INTO public.admin_users (first_name, last_name, email, password)
		VALUES ($1, $2, $3, $4)
	`

	// Use request-scoped context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err = dbPool.Exec(ctx, query, reg.FirstName, reg.LastName, reg.Email, string(hashedPassword))
	if err != nil {
		// Handle duplicate email separately
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			http.Error(w, "Email already registered", http.StatusConflict) // 409
			log.Println("[Signup] Duplicate email:", reg.Email)
			return
		}

		// Handle broken pipe / transient DB errors
		if err.Error() == "broken pipe" {
			http.Error(w, "Temporary database error, try again", http.StatusServiceUnavailable)
			log.Println("[Signup] Broken pipe, database connection dropped")
			return
		}

		http.Error(w, "Database insert failed: "+err.Error(), http.StatusInternalServerError)
		log.Println("[Signup] Database insert error:", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User " + reg.Email + " registered successfully!"))
	log.Println("[Signup] User registered:", reg.Email)
}

// --- LoginHandler ---
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	if dbPool == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		log.Println("[Login] DB pool is nil")
		return
	}

	var creds Login
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		log.Println("[Login] JSON decode error:", err)
		return
	}

	var storedHashedPassword string

	// Add context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err := dbPool.QueryRow(ctx,
		"SELECT password FROM public.admin_users WHERE email=$1",
		creds.Email).Scan(&storedHashedPassword)

	if err != nil {
		log.Println("[Login] Query error:", err)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// ✅ Compare hashed password
	if bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(creds.Password)) != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	w.Write([]byte("User " + creds.Email + " logged in!"))
	log.Println("[Login] User logged in:", creds.Email)
}

// --- Register routes ---
func RegisterAuthRoutes(r *mux.Router) {
	r.HandleFunc("/register", SignupHandler).Methods("POST")
	r.HandleFunc("/login", LoginHandler).Methods("POST")
}




