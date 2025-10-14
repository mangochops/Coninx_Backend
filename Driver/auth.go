package Driver

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Driver struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	IDNumber    int    `json:"idNumber"`
	Password    string `json:"password"`
	PhoneNumber *int64 `json:"phoneNumber,omitempty"` // nullable
}

var db *pgxpool.Pool

// InitDB sets the global DB connection
func InitDB(pool *pgxpool.Pool) {
	db = pool
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var d Driver
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	// Insert NULL if phone number is not provided
	_, err := db.Exec(r.Context(),
		"INSERT INTO drivers (first_name, last_name, id_number, password, phone_number) VALUES ($1, $2, $3, $4, $5)",
		d.FirstName, d.LastName, d.IDNumber, d.Password, d.PhoneNumber,
	)
	if err != nil {
		http.Error(w, "Failed to register driver: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Driver " + d.FirstName + " " + d.LastName + " registered successfully!"))
}

func GetDriversHandler(w http.ResponseWriter, r *http.Request) {
	if db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	rows, err := db.Query(r.Context(),
		"SELECT id_number, first_name, last_name, phone_number FROM drivers")
	if err != nil {
		http.Error(w, "Failed to fetch drivers: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var drivers []Driver
	for rows.Next() {
		var d Driver
		if err := rows.Scan(&d.IDNumber, &d.FirstName, &d.LastName, &d.PhoneNumber); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		drivers = append(drivers, d)
	}

	json.NewEncoder(w).Encode(drivers)
}

func GetDriverByIDHandler(w http.ResponseWriter, r *http.Request) {
	if db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid driver ID", http.StatusBadRequest)
		return
	}

	var d Driver
	err = db.QueryRow(r.Context(),
		"SELECT id_number, first_name, last_name, phone_number FROM drivers WHERE id_number=$1", id,
	).Scan(&d.IDNumber, &d.FirstName, &d.LastName, &d.PhoneNumber)

	if err != nil {
		http.Error(w, "Driver not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(d)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	// Accept IDNumber as string to match frontend input
	var creds struct {
		IDNumber string `json:"idNumber"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Convert IDNumber to int
	idNum, err := strconv.Atoi(creds.IDNumber)
	if err != nil {
		http.Error(w, "Invalid ID number", http.StatusBadRequest)
		return
	}

	// Check DB for driver
	var storedPassword string
	err = db.QueryRow(r.Context(),
		"SELECT password FROM drivers WHERE id_number=$1", idNum,
	).Scan(&storedPassword)
	if err != nil {
		http.Error(w, "Driver not found", http.StatusNotFound)
		return
	}

	// Check password
	if storedPassword != creds.Password {
		http.Error(w, "Incorrect password", http.StatusUnauthorized)
		return
	}

	// Success — return JSON
	resp := map[string]interface{}{
		"driverId": idNum,
		"message":  "Login successful",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// RegisterDriverRoutes registers the driver endpoints to the router
func RegisterDriverRoutes(r *mux.Router) {
	r.HandleFunc("/register", RegisterHandler).Methods("POST")
	r.HandleFunc("/login", LoginHandler).Methods("POST")
	r.HandleFunc("/all", GetDriversHandler).Methods("GET")
	r.HandleFunc("/{id}", GetDriverByIDHandler).Methods("GET")
}
