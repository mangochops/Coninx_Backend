package Driver

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Delivery represents proof of completed dispatch by driver
type Delivery struct {
	ID         int `json:"id"`
	DispatchID int `json:"dispatchId"`

	Date   time.Time `json:"date"`
	TripID int       `json:"tripId"`
}

// ---------------- CRUD ----------------

// CreateDeliveryHandler inserts a new delivery record
func CreateDeliveryHandler(w http.ResponseWriter, r *http.Request) {
	var d Delivery
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Check if trip exists in admin.trips
	var tripExists bool
	err := db.QueryRow(
		context.Background(),
		`SELECT EXISTS(SELECT 1 FROM trips WHERE id=$1)`,
		d.TripID,
	).Scan(&tripExists)
	if err != nil {
		http.Error(w, "Failed to check trip: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if !tripExists {
		http.Error(w, "Trip not found", http.StatusBadRequest)
		return
	}

	err = db.QueryRow(
		context.Background(),
		`INSERT INTO deliveries (dispatch_id, trip_id, date)
		 VALUES ($1, $2, NOW())
		 RETURNING id, date`,
		d.DispatchID, d.TripID,
	).Scan(&d.ID, &d.Date)
	if err != nil {
		http.Error(w, "Failed to insert delivery: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(d)
}

// ListDeliveriesHandler returns all deliveries
func ListDeliveriesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(
		context.Background(),
		`SELECT id, dispatch_id, trip_id, date
		 FROM deliveries
		 ORDER BY date DESC`,
	)
	if err != nil {
		http.Error(w, "Failed to fetch deliveries: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []Delivery
	for rows.Next() {
		var d Delivery
		if err := rows.Scan(&d.ID, &d.DispatchID, &d.TripID, &d.Date); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, d)
	}

	json.NewEncoder(w).Encode(list)
}

// GetDeliveryHandler returns a single delivery by ID
func GetDeliveryHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var d Delivery
	err := db.QueryRow(
		context.Background(),
		`SELECT id, dispatch_id, trip_id, date
		 FROM deliveries WHERE id=$1`,
		id,
	).Scan(&d.ID, &d.DispatchID, &d.TripID, &d.Date)
	if err != nil {
		http.Error(w, "Delivery not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(d)
}

// DeleteDeliveryHandler deletes a delivery by ID
func DeleteDeliveryHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	_, err := db.Exec(context.Background(),
		`DELETE FROM deliveries WHERE id=$1`, id,
	)
	if err != nil {
		http.Error(w, "Failed to delete delivery: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ---------------- ROUTES ----------------

// RegisterDeliveryRoutes adds delivery endpoints to router
func RegisterDeliveryRoutes(r *mux.Router) {
	r.HandleFunc("/driver/deliveries", CreateDeliveryHandler).Methods("POST")
	r.HandleFunc("/driver/deliveries", ListDeliveriesHandler).Methods("GET")
	r.HandleFunc("/driver/deliveries/{id}", GetDeliveryHandler).Methods("GET")
	r.HandleFunc("/driver/deliveries/{id}", DeleteDeliveryHandler).Methods("DELETE")
}
