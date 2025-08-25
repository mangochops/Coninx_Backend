package Driver

import (
	"encoding/json"
	
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Trip struct {
	ID            int         `json:"id"`
	Destination   string      `json:"destination"`
	DriverID      int         `json:"driverId"`
	RecipientName string      `json:"recipientName"`
	Invoice       int         `json:"invoice"`
	Status        string      `json:"status"` // e.g., "requested", "in-progress", "completed"
	Coordinates   Coordinates `json:"coordinates"`
}

var (
	trips   = make(map[int]Trip)
	nextID  = 1
	tripsMu sync.Mutex
)

// Create a new trip
func CreateTripHandler(w http.ResponseWriter, r *http.Request) {
	var t Trip
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	tripsMu.Lock()
	t.ID = nextID
	nextID++
	trips[t.ID] = t
	tripsMu.Unlock()
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

// Get trip info by ID
func GetTripHandler(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}
	tripsMu.Lock()
	trip, ok := trips[id]
	tripsMu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	json.NewEncoder(w).Encode(trip)
}

// Update trip coordinates
func UpdateTripCoordinatesHandler(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}
	var coords Coordinates
	if err := json.NewDecoder(r.Body).Decode(&coords); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	tripsMu.Lock()
	trip, ok := trips[id]
	if ok {
		trip.Coordinates = coords
		trips[id] = trip
	}
	tripsMu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	json.NewEncoder(w).Encode(trip)
}

// Update trip status
func UpdateTripStatusHandler(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}
	var status struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	tripsMu.Lock()
	trip, ok := trips[id]
	if ok {
		trip.Status = status.Status
		trips[id] = trip
	}
	tripsMu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	json.NewEncoder(w).Encode(trip)
}

// RegisterTripRoutes registers all trip endpoints to the router
func RegisterTripRoutes(r *mux.Router) {
	r.HandleFunc("/trip", CreateTripHandler).Methods("POST")
	r.HandleFunc("/trip/{id}", GetTripHandler).Methods("GET")
	r.HandleFunc("/trip/{id}/coordinates", UpdateTripCoordinatesHandler).Methods("PUT")
	r.HandleFunc("/trip/{id}/status", UpdateTripStatusHandler).Methods("PUT")
}


