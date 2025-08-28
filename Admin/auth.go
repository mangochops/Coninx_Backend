package Admin

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

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

var db *pgx.Conn

// InitDB sets the global DB connection
func InitDB(conn *pgx.Conn) {
	db = conn
}

// SignupHandler handles admin registration
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	if db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		log.Println("[Signup] DB connection is nil")
		return
	}

	var reg Register
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		log.Println("[Signup] JSON decode error:", err)
		return
	}

	

	query := `
		INSERT INTO public.admin_users (first_name, last_name, email, password)
		VALUES ($1, $2, $3, $4)
	`

	_, err := db.Exec(context.Background(), query, reg.FirstName, reg.LastName, reg.Email, reg.Password)
	if err != nil {
		log.Println("[Signup] Database insert error:", err)
		http.Error(w, "Database insert failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User " + reg.Email + " registered successfully!"))
	log.Println("[Signup] User registered:", reg.Email)
}

// LoginHandler handles admin login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	if db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		log.Println("[Login] DB connection is nil")
		return
	}

	var creds Login
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		log.Println("[Login] JSON decode error:", err)
		return
	}

	var storedPassword string
	err := db.QueryRow(context.Background(),
		"SELECT password FROM public.admin_users WHERE email=$1",
		creds.Email).Scan(&storedPassword)

	if err != nil {
		log.Println("[Login] Query error:", err)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if creds.Password != storedPassword {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	w.Write([]byte("User " + creds.Email + " logged in!"))
	log.Println("[Login] User logged in:", creds.Email)
}

// RegisterAuthRoutes registers /admin/register and /admin/login
func RegisterAuthRoutes(r *mux.Router) {
	r.HandleFunc("/register", SignupHandler).Methods("POST")
	r.HandleFunc("/login", LoginHandler).Methods("POST")
}


