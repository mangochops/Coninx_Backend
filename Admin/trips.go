package Admin

import (
	"context"
	"encoding/json"

	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/mangochops/coninx_backend/Driver"
)

// Trips represents a delivery trip persisted in DB
type Trips struct {
	ID            int           `json:"id"`
	DispatchID    int           `json:"dispatch_id"`
	Dispatch      *Dispatch     `json:"dispatch,omitempty"`
	Driver        Driver.Driver `json:"driver"`
	Vehicle       Vehicle       `json:"vehicle"`
	Destination   string        `json:"destination"`
	RecipientName string        `json:"recipient_name"`
	Status        string        `json:"status"`
	Latitude      float64       `json:"latitude"`
	Longitude     float64       `json:"longitude"`
	LastUpdated   time.Time     `json:"lastUpdated"`
}

// ---------------- SSE broadcaster ----------------

type sseClient chan []byte

var (
	sseClients   = make(map[sseClient]bool)
	sseClientsMu sync.Mutex
)

// broadcastToSSE sends the given payload to all connected SSE clients.
func broadcastToSSE(payload interface{}) {
	b, err := json.Marshal(payload)
	if err != nil {
		log.Println("broadcast marshal error:", err)
		return
	}

	sseClientsMu.Lock()
	defer sseClientsMu.Unlock()
	for ch := range sseClients {
		select {
		case ch <- b:
		default:
			// Drop if client is too slow
		}
	}
}

// sseHandler handles new SSE client connections.

// ---------------- CRUD ----------------

// AutoCreateTrip is called by CreateDispatch to attach a trip automatically
func AutoCreateTrip(dispatchID int, driverID int, vehicleID int, destination string, recipientName string) (*Trips, error) {
	var t Trips
	err := dbPool.QueryRow(
		context.Background(),
		`INSERT INTO trips (dispatch_id, driver_id, vehicle_id, destination, recipient_name, status, latitude, longitude, last_updated)
		 VALUES ($1, $2, $3, $4, $5, 'started', 0, 0, NOW())
		 RETURNING id, dispatch_id, driver_id, vehicle_id, destination, recipient_name, status, latitude, longitude, last_updated`,
		dispatchID, driverID, vehicleID, destination, recipientName,
	).Scan(
		&t.ID,
		&t.DispatchID,
		&t.Driver.IDNumber,
		&t.Vehicle.ID,
		&t.Destination,
		&t.RecipientName,
		&t.Status,
		&t.Latitude,
		&t.Longitude,
		&t.LastUpdated,
	)
	if err != nil {
		return nil, err
	}

	// broadcast new trip
	broadcastToSSE(map[string]interface{}{
		"type": "trip_created",
		"trip": t,
	})

	return &t, nil
}

func GetTrips(w http.ResponseWriter, r *http.Request) {
	rows, err := dbPool.Query(context.Background(),
		`SELECT id, dispatch_id, driver_id, vehicle_id, status, latitude, longitude, last_updated 
		 FROM trips WHERE status != 'completed'`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var res []Trips
	for rows.Next() {
		var t Trips
		if err := rows.Scan(&t.ID, &t.DispatchID, &t.Driver.IDNumber, &t.Vehicle.ID,
			&t.Status, &t.Latitude, &t.Longitude, &t.LastUpdated); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		res = append(res, t)
	}

	json.NewEncoder(w).Encode(res)
}

func GetTrip(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var t Trips
	err = dbPool.QueryRow(context.Background(),
		`SELECT id, dispatch_id, driver_id, vehicle_id, status, latitude, longitude, last_updated
		 FROM trips WHERE id=$1`, id,
	).Scan(&t.ID, &t.DispatchID, &t.Driver.IDNumber, &t.Vehicle.ID,
		&t.Status, &t.Latitude, &t.Longitude, &t.LastUpdated)

	if err != nil {
		http.Error(w, "Trip not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(t)
}

// Get trips assigned to a specific driver
func GetTripsByDriver(w http.ResponseWriter, r *http.Request) {
	driverIDStr := mux.Vars(r)["driverId"]
	driverID, err := strconv.Atoi(driverIDStr)
	if err != nil {
		http.Error(w, "Invalid driver ID", http.StatusBadRequest)
		return
	}

	rows, err := dbPool.Query(context.Background(),
		`SELECT id, dispatch_id, driver_id, vehicle_id, status, latitude, longitude, last_updated
		 FROM trips WHERE driver_id=$1`, driverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var res []Trips
	for rows.Next() {
		var t Trips
		if err := rows.Scan(&t.ID, &t.DispatchID, &t.Driver.IDNumber, &t.Vehicle.ID,
			&t.Status, &t.Latitude, &t.Longitude, &t.LastUpdated); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		res = append(res, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
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

	_, err = dbPool.Exec(context.Background(),
		`UPDATE trips SET status=$1, latitude=$2, longitude=$3, last_updated=NOW()
		 WHERE id=$4`,
		updated.Status, updated.Latitude, updated.Longitude, id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updated.ID = id
	updated.LastUpdated = time.Now()
	json.NewEncoder(w).Encode(updated)

	broadcastToSSE(map[string]interface{}{
		"type": "trip_updated",
		"trip": updated,
	})
}

func DeleteTrip(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	_, err := dbPool.Exec(context.Background(), `DELETE FROM trips WHERE id=$1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	broadcastToSSE(map[string]interface{}{
		"type":   "trip_deleted",
		"tripId": id,
	})
}

// ---------------- Tracking ----------------

func UpdateTripLocation(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	var body struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	_, err := dbPool.Exec(context.Background(),
		`UPDATE trips SET latitude=$1, longitude=$2, last_updated=NOW() WHERE id=$3`,
		body.Latitude, body.Longitude, id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var t Trips
	err = dbPool.QueryRow(context.Background(),
		`SELECT id, dispatch_id, driver_id, vehicle_id, status, latitude, longitude, last_updated 
		 FROM trips WHERE id=$1 AND status != 'completed'`, id,
	).Scan(&t.ID, &t.DispatchID, &t.Driver.IDNumber, &t.Vehicle.ID,
		&t.Status, &t.Latitude, &t.Longitude, &t.LastUpdated)
	if err != nil {
		http.Error(w, "Trip not found or already completed", http.StatusNotFound)
		return
	}

	broadcastToSSE(map[string]interface{}{
		"type": "location_update",
		"trip": t,
	})

	json.NewEncoder(w).Encode(t)
}

// ---------------- New Endpoints ----------------

// GetTripsByDispatch fetches all trips for a given dispatch
func GetTripsByDispatch(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["dispatchId"]
	dispatchID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Dispatch ID", http.StatusBadRequest)
		return
	}

	rows, err := dbPool.Query(context.Background(),
		`SELECT id, dispatch_id, driver_id, vehicle_id, status, latitude, longitude, last_updated
		 FROM trips WHERE dispatch_id=$1`, dispatchID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var res []Trips
	for rows.Next() {
		var t Trips
		if err := rows.Scan(&t.ID, &t.DispatchID, &t.Driver.IDNumber, &t.Vehicle.ID,
			&t.Status, &t.Latitude, &t.Longitude, &t.LastUpdated); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		res = append(res, t)
	}

	json.NewEncoder(w).Encode(res)
}

// CompleteTrip marks a trip as completed
func CompleteTrip(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	_, err := dbPool.Exec(context.Background(),
		`UPDATE trips SET status='completed', last_updated=NOW() WHERE id=$1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	broadcastToSSE(map[string]interface{}{
		"type":   "trip_completed",
		"tripId": id,
	})

	w.WriteHeader(http.StatusNoContent)
}

// RegisterTripRoutes registers all trip endpoints
func RegisterTripRoutes(r *mux.Router) {
	r.HandleFunc("/trips", GetTrips).Methods("GET")
	r.HandleFunc("/trips/{id}", GetTrip).Methods("GET")
	r.HandleFunc("/trips/{id}", UpdateTrip).Methods("PUT")
	r.HandleFunc("/trips/{id}", DeleteTrip).Methods("DELETE")
	r.HandleFunc("/trips/{id}/complete", CompleteTrip).Methods("PUT")

	// Fetch trips by dispatch
	r.HandleFunc("/dispatches/{dispatchId}/trips", GetTripsByDispatch).Methods("GET")

	// Live tracking
	r.HandleFunc("/trips/{id}/location", UpdateTripLocation).Methods("PUT")

}
