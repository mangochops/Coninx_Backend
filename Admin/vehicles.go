package Admin

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Vehicle struct
type Vehicle struct {
	ID     int    `json:"id"`
	Type   string `json:"type"`
	RegNo  string `json:"reg_no"`
	Status bool   `json:"status"`
}

// CreateVehicle inserts a new vehicle
func CreateVehicle(w http.ResponseWriter, r *http.Request) {
	var v Vehicle
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Use db from auth.go
	err := dbPool.QueryRow(
		context.Background(),
		`INSERT INTO vehicles (type, reg_no, status) VALUES ($1, $2, $3) RETURNING id`,
		v.Type, v.RegNo, v.Status,
	).Scan(&v.ID)
	if err != nil {
		http.Error(w, "Failed to insert vehicle: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(v)
}

// GetVehicles returns all vehicles
func GetVehicles(w http.ResponseWriter, r *http.Request) {
	rows, err := dbPool.Query(context.Background(), `SELECT id, type, reg_no, status FROM vehicles`)
	if err != nil {
		http.Error(w, "Failed to fetch vehicles: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var vehicles []Vehicle
	for rows.Next() {
		var v Vehicle
		if err := rows.Scan(&v.ID, &v.Type, &v.RegNo, &v.Status); err != nil {
			http.Error(w, "Error scanning vehicle: "+err.Error(), http.StatusInternalServerError)
			return
		}
		vehicles = append(vehicles, v)
	}

	json.NewEncoder(w).Encode(vehicles)
}

// GetVehicle returns a single vehicle by ID
func GetVehicle(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var v Vehicle
	err = dbPool.QueryRow(context.Background(),
		`SELECT id, type, reg_no, status FROM vehicles WHERE id = $1`,
		id,
	).Scan(&v.ID, &v.Type, &v.RegNo, &v.Status)

	if err != nil {
		// Handle "no rows" without pgx.ErrNoRows directly
		if err.Error() == "no rows in result set" {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Failed to fetch vehicle: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(v)
}

// UpdateVehicle updates a vehicle by ID
func UpdateVehicle(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var v Vehicle
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	_, err = dbPool.Exec(context.Background(),
		`UPDATE vehicles SET type=$1, reg_no=$2, status=$3 WHERE id=$4`,
		v.Type, v.RegNo, v.Status, id,
	)
	if err != nil {
		http.Error(w, "Failed to update vehicle: "+err.Error(), http.StatusInternalServerError)
		return
	}

	v.ID = id
	json.NewEncoder(w).Encode(v)
}

// DeleteVehicle removes a vehicle by ID
func DeleteVehicle(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = dbPool.Exec(context.Background(), `DELETE FROM vehicles WHERE id=$1`, id)
	if err != nil {
		http.Error(w, "Failed to delete vehicle: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RegisterVehicleRoutes registers vehicle endpoints
func RegisterVehicleRoutes(r *mux.Router) {
	r.HandleFunc("/vehicles", CreateVehicle).Methods("POST")
	r.HandleFunc("/vehicles", GetVehicles).Methods("GET")
	r.HandleFunc("/vehicles/{id}", GetVehicle).Methods("GET")
	r.HandleFunc("/vehicles/{id}", UpdateVehicle).Methods("PUT")
	r.HandleFunc("/vehicles/{id}", DeleteVehicle).Methods("DELETE")
}

