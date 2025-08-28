package Admin

import (
	"context"
	"encoding/json"
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

var db *pgx.Conn // global connection (set from main.go)

// InitDB allows main.go to pass in the pgx connection
func InitDB(conn *pgx.Conn) {
	db = conn
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var reg Register
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Insert into admin_users table
	query := `
		INSERT INTO admin_users (username, email, password, created_at)
		VALUES ($1, $2, $3, NOW())
	`
	username := reg.FirstName + " " + reg.LastName
	_, err := db.Exec(context.Background(), query, username, reg.Email, reg.Password)
	if err != nil {
		http.Error(w, "Database insert failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User " + reg.Email + " registered successfully!"))
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// TODO: Replace with proper password check
	var storedPassword string
	err := db.QueryRow(context.Background(),
		"SELECT password FROM admin_users WHERE email=$1",
		creds.Email).Scan(&storedPassword)

	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if creds.Password != storedPassword {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	w.Write([]byte("User " + creds.Email + " logged in!"))
}

// RegisterAuthRoutes registers the auth endpoints to the router
func RegisterAuthRoutes(r *mux.Router) {
	r.HandleFunc("/register", SignupHandler).Methods("POST")
	r.HandleFunc("/login", LoginHandler).Methods("POST")
}

