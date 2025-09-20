package Driver

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

type Driver struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	IDNumber  int    `json:"idNumber"`
	Password  string `json:"password"`
}

var drivers []Driver
var db *pgx.Conn

// InitDB sets the global DB connection
func InitDB(conn *pgx.Conn) {
	db = conn
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
	// Save driver to database
	if db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}
	_, err := db.Exec(r.Context(),
		"INSERT INTO drivers (first_name, last_name, id_number, password) VALUES ($1, $2, $3, $4)",
		d.FirstName, d.LastName, d.IDNumber, d.Password,
	)
	if err != nil {
		http.Error(w, "Failed to register driver: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Driver " + d.FirstName + " " + d.LastName + " registered successfully!"))
}

func GetDriver(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(drivers)
}

// Optionally, add a login handler for drivers
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds struct {
		IDNumber int    `json:"idNumber"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// TODO: Authenticate driver with database
	w.Write([]byte("Driver " + strconv.Itoa(creds.IDNumber) + " logged in!"))
}

func GetDriversHandler(w http.ResponseWriter, r *http.Request) {
	if db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	rows, err := db.Query(r.Context(), "SELECT id_number, first_name, last_name FROM drivers")
	if err != nil {
		http.Error(w, "Failed to fetch drivers: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var drivers []Driver
	for rows.Next() {
		var d Driver
		if err := rows.Scan(&d.IDNumber, &d.FirstName, &d.LastName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		drivers = append(drivers, d)
	}

	json.NewEncoder(w).Encode(drivers)
}

// GetDriverByIDHandler fetches a single driver by IDNumber
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
		"SELECT id_number, first_name, last_name FROM drivers WHERE id_number=$1", id,
	).Scan(&d.IDNumber, &d.FirstName, &d.LastName)

	if err != nil {
		http.Error(w, "Driver not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(d)
}

// RegisterDriverRoutes registers the driver endpoints to the router
func RegisterDriverRoutes(r *mux.Router) {
	r.HandleFunc("/register", RegisterHandler).Methods("POST")
	r.HandleFunc("/login", LoginHandler).Methods("POST")
	r.HandleFunc("/all", GetDriversHandler).Methods("GET")
	r.HandleFunc("/{id}", GetDriverByIDHandler).Methods("GET")
}
