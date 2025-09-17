package Admin

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgconn"
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

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// --- SignupHandler ---
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "Only POST allowed"})
		return
	}

	if dbPool == nil {
		http.Error(w, `{"success":false,"message":"Database not initialized"}`, http.StatusInternalServerError)
		log.Println("[Signup] DB pool is nil")
		return
	}

	var reg Register
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		http.Error(w, `{"success":false,"message":"Invalid input"}`, http.StatusBadRequest)
		log.Println("[Signup] JSON decode error:", err)
		return
	}

	// ✅ Hash password before saving
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reg.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"success":false,"message":"Error securing password"}`, http.StatusInternalServerError)
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
			http.Error(w, `{"success":false,"message":"Email already registered"}`, http.StatusConflict)
			log.Println("[Signup] Duplicate email:", reg.Email)
			return
		}

		http.Error(w, `{"success":false,"message":"Database insert failed"}`, http.StatusInternalServerError)
		log.Println("[Signup] Database insert error:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Message: "User registered successfully!",
		Data: map[string]string{
			"firstName": reg.FirstName,
			"lastName":  reg.LastName,
			"email":     reg.Email,
		},
	})

	log.Println("[Signup] User registered:", reg.Email)
}

// --- LoginHandler ---
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "Only POST allowed"})
		return
	}

	if dbPool == nil {
		http.Error(w, `{"success":false,"message":"Database not initialized"}`, http.StatusInternalServerError)
		log.Println("[Login] DB pool is nil")
		return
	}

	var creds Login
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, `{"success":false,"message":"Invalid input"}`, http.StatusBadRequest)
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
		http.Error(w, `{"success":false,"message":"Invalid email or password"}`, http.StatusUnauthorized)
		log.Println("[Login] Query error:", err)
		return
	}

	// ✅ Compare hashed password
	if bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(creds.Password)) != nil {
		http.Error(w, `{"success":false,"message":"Invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Message: "User logged in!",
		Data: map[string]string{
			"email": creds.Email,
		},
	})

	log.Println("[Login] User logged in:", creds.Email)
}

// --- Register routes ---
func RegisterAuthRoutes(r *mux.Router) {
	r.HandleFunc("/register", SignupHandler).Methods("POST")
	r.HandleFunc("/login", LoginHandler).Methods("POST")
}





