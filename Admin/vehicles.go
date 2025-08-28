package Admin

import (
	"encoding/json"
	
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Vehicle struct {
	ID        int       `json:"id"`
	Type      string    `json:"type"`
	RegNo     string    `json:"reg_no"`
	Status    bool      `json:"status"`
}

var vehicles []Vehicle


func CreateVehicle(w http.ResponseWriter, r *http.Request) {
	var v Vehicle
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	v.ID = nextID
	nextID++
	vehicles = append(vehicles, v)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(v)
}

func GetVehicles(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(vehicles)
}

func GetVehicle(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	for _, v := range vehicles {
		if v.ID == id {
			json.NewEncoder(w).Encode(v)
			return
		}
	}
	http.NotFound(w, r)
}

func UpdateVehicle(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	var updated Vehicle
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	for i, v := range vehicles {
		if v.ID == id {
			updated.ID = id
			vehicles[i] = updated
			json.NewEncoder(w).Encode(updated)
			return
		}
	}
	http.NotFound(w, r)
}

func DeleteVehicle(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	for i, v := range vehicles {
		if v.ID == id {
			vehicles = append(vehicles[:i], vehicles[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.NotFound(w, r)
}

// RegisterVehicleRoutes registers all vehicle endpoints to the router
func RegisterVehicleRoutes(r *mux.Router) {
	r.HandleFunc("/vehicles", CreateVehicle).Methods("POST")
	r.HandleFunc("/vehicles", GetVehicles).Methods("GET")
	r.HandleFunc("/vehicles/{id}", GetVehicle).Methods("GET")
	r.HandleFunc("/vehicles/{id}", UpdateVehicle).Methods("PUT")
	r.HandleFunc("/vehicles/{id}", DeleteVehicle).Methods("DELETE")
}