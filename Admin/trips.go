package Admin

import (
	"encoding/json"
	
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mangochops/coninx_backend/Driver"

)

type Trips struct {
	ID        int       `json:"id"`
	Dispatch  Dispatch  `json:"dispatch"`
	Driver    Driver.Driver `json:"driver"`
	Vehicle   Vehicle   `json:"vehicle"`
	Status    string    `json:"status"`

}

var trips []Trips


	func CreateTrip(w http.ResponseWriter, r *http.Request) {
		var t Trips
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	t.ID = nextID
	nextID++
	trips = append(trips, t)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func GetTrips(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(trips)
}

func GetTrip(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	for _, t := range trips {
		if t.ID == id {
			json.NewEncoder(w).Encode(t)
			return
		}
	}
	http.NotFound(w, r)
}

func UpdateTrip(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	var updated Trips
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	for i, t := range trips {
		if t.ID == id {
			updated.ID = id
			trips[i] = updated
			json.NewEncoder(w).Encode(updated)
			return
		}
	}
	http.NotFound(w, r)
}

func DeleteTrip(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	for i, t := range trips {
		if t.ID == id {
			trips = append(trips[:i], trips[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.NotFound(w, r)
}

// RegisterTripRoutes registers all trip endpoints to the router
func RegisterTripRoutes(r *mux.Router) {
	r.HandleFunc("/trips", CreateTrip).Methods("POST")
	r.HandleFunc("/trips", GetTrips).Methods("GET")
	r.HandleFunc("/trips/{id}", GetTrip).Methods("GET")
	r.HandleFunc("/trips/{id}", UpdateTrip).Methods("PUT")
	r.HandleFunc("/trips/{id}", DeleteTrip).Methods("DELETE")
}