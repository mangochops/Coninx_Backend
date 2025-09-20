package Admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/mangochops/coninx_backend/Driver"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/verify/v2"
)

type Dispatch struct {
	ID        int    `json:"id"`
	Recipient string `json:"recipient"` // ✅ corrected

	Location string        `json:"location"`
	Driver   Driver.Driver `json:"driver"`
	Vehicle  Vehicle       `json:"vehicle"`
	Invoice  int           `json:"invoice"`
	Date     time.Time     `json:"date"`
	Verified bool          `json:"verified"`
}

// var db *pgxpool.Pool

// Twilio vars
var (
	accountSid string
	authToken  string
	serviceSid string
	client     *twilio.RestClient
)

// func InitDB() {
// 	dsn := os.Getenv("DATABASE_URL")
// 	var err error
// 	db, err = pgxpool.New(context.Background(), dsn)
// 	if err != nil {
// 		log.Fatalf("Unable to connect to database: %v", err)
// 	}
// }

// Init Twilio + DB
func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system env")
	}

	accountSid = os.Getenv("TWILIO_ACCOUNT_SID")
	authToken = os.Getenv("TWILIO_AUTH_TOKEN")
	serviceSid = os.Getenv("TWILIO_VERIFY_SERVICE_SID")

	client = twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})

	// InitDB()
}

// Create a dispatch
func CreateDispatch(w http.ResponseWriter, r *http.Request) {
	var d Dispatch
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// 🔑 Lookup driver_id by driver.id_number
	var driverID int
	err := dbPool.QueryRow(
		context.Background(),
		"SELECT id FROM drivers WHERE id_number=$1",
		d.Driver.IDNumber,
	).Scan(&driverID)
	if err != nil {
		http.Error(w, "Driver not found", http.StatusBadRequest)
		return
	}

	// 🔑 Lookup vehicle_id by vehicle.reg_no
	var vehicleID int
	err = dbPool.QueryRow(
		context.Background(),
		"SELECT id FROM vehicles WHERE reg_no=$1",
		d.Vehicle.RegNo,
	).Scan(&vehicleID)
	if err != nil {
		http.Error(w, "Vehicle not found", http.StatusBadRequest)
		return
	}

	// Insert dispatch
	err = dbPool.QueryRow(
		context.Background(),
		`INSERT INTO dispatches (recipient, location, driver_id, vehicle_id, invoice, verified, date)
         VALUES ($1, $2, $3, $4, $5, FALSE, NOW())
         RETURNING id, date, verified`,
		d.Recipient, d.Location, driverID, vehicleID, d.Invoice,
	).Scan(&d.ID, &d.Date, &d.Verified)
	if err != nil {
		http.Error(w, "Error inserting dispatch", http.StatusInternalServerError)
		return
	}

	// 🔑 Auto-create a trip for this dispatch
	trip, err := AutoCreateTrip(d.ID, driverID, vehicleID, d.Location, d.Recipient)
	if err != nil {
		http.Error(w, "Dispatch created but trip creation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with both Dispatch and Trip
	response := map[string]interface{}{
		"dispatch": d,
		"trip":     trip,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Get all dispatches (with driver + vehicle info)
func GetDispatches(w http.ResponseWriter, r *http.Request) {
	rows, err := dbPool.Query(context.Background(), `
		SELECT d.id, d.recipient, d.location, d.invoice, d.date, d.verified,
		       dr.id_number, dr.first_name || ' ' || dr.last_name AS driver_name,
		       v.reg_no
		FROM dispatches d
		LEFT JOIN drivers dr ON d.driver_id = dr.id
		LEFT JOIN vehicles v ON d.vehicle_id = v.id
		ORDER BY d.date DESC
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var dispatches []Dispatch
	for rows.Next() {
		var d Dispatch
		var driverIDNumber sql.NullInt64
		var driverName sql.NullString
		var vehicleReg sql.NullString

		if err := rows.Scan(&d.ID, &d.Recipient, &d.Location, &d.Invoice, &d.Date, &d.Verified,
			&driverIDNumber, &driverName, &vehicleReg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if driverIDNumber.Valid {
			d.Driver.IDNumber = int(driverIDNumber.Int64)
		}
		if driverName.Valid {
			d.Driver.FirstName = driverName.String
		}
		if vehicleReg.Valid {
			d.Vehicle.RegNo = vehicleReg.String
		}

		dispatches = append(dispatches, d)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dispatches)
}

// Get single dispatch by ID
func GetDispatch(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	var d Dispatch
	var driverIDNumber sql.NullInt64
	var driverName sql.NullString
	var vehicleReg sql.NullString

	err := dbPool.QueryRow(context.Background(), `
		SELECT d.id, d.recipient, d.location, d.invoice, d.date, d.verified,
		       dr.id_number, dr.first_name || ' ' || dr.last_name AS driver_name,
		       v.reg_no
		FROM dispatches d
		LEFT JOIN drivers dr ON d.driver_id = dr.id
		LEFT JOIN vehicles v ON d.vehicle_id = v.id
		WHERE d.id=$1
	`, id).Scan(&d.ID, &d.Recipient, &d.Location, &d.Invoice, &d.Date, &d.Verified,
		&driverIDNumber, &driverName, &vehicleReg)

	if err != nil {
		http.Error(w, "Dispatch not found", http.StatusNotFound)
		return
	}

	if driverIDNumber.Valid {
		d.Driver.IDNumber = int(driverIDNumber.Int64)
	}
	if driverName.Valid {
		d.Driver.FirstName = driverName.String
	}
	if vehicleReg.Valid {
		d.Vehicle.RegNo = vehicleReg.String
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d)
}

// Update dispatch (recipient, location, invoice, driver, vehicle)
func UpdateDispatch(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	var updated Dispatch
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Lookup driver ID
	var driverID int
	err := dbPool.QueryRow(context.Background(),
		"SELECT id FROM drivers WHERE id_number=$1", updated.Driver.IDNumber,
	).Scan(&driverID)
	if err != nil {
		http.Error(w, "Driver not found", http.StatusBadRequest)
		return
	}

	// Lookup vehicle ID
	var vehicleID int
	err = dbPool.QueryRow(context.Background(),
		"SELECT id FROM vehicles WHERE reg_no=$1", updated.Vehicle.RegNo,
	).Scan(&vehicleID)
	if err != nil {
		http.Error(w, "Vehicle not found", http.StatusBadRequest)
		return
	}

	// Update dispatch
	_, err = dbPool.Exec(context.Background(),
		`UPDATE dispatches 
         SET recipient=$1, location=$2, invoice=$3, driver_id=$4, vehicle_id=$5
         WHERE id=$6`,
		updated.Recipient, updated.Location, updated.Invoice, driverID, vehicleID, id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updated.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// Delete dispatch
func DeleteDispatch(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	_, err := dbPool.Exec(context.Background(), `DELETE FROM dispatches WHERE id=$1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------------- OTP Endpoints ----------------

func SendOTP(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	var phone string
	err := dbPool.QueryRow(context.Background(),
		`SELECT phone FROM dispatches WHERE id=$1`, id).Scan(&phone)
	if err != nil {
		http.Error(w, "Dispatch not found", http.StatusNotFound)
		return
	}

	params := &openapi.CreateVerificationParams{}
	params.SetTo(phone)
	params.SetChannel("sms")

	resp, err := client.VerifyV2.CreateVerification(serviceSid, params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": *resp.Status,
	})
}

func VerifyOTP(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	dispatchID, _ := strconv.Atoi(idStr)

	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Get phone number
	var phone string
	err := dbPool.QueryRow(context.Background(),
		`SELECT phone FROM dispatches WHERE id=$1`, dispatchID).Scan(&phone)
	if err != nil {
		http.Error(w, "Dispatch not found", http.StatusNotFound)
		return
	}

	// Verify OTP with Twilio
	params := &openapi.CreateVerificationCheckParams{}
	params.SetTo(phone)
	params.SetCode(body.Code)

	resp, err := client.VerifyV2.CreateVerificationCheck(serviceSid, params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if *resp.Status != "approved" {
		http.Error(w, "Invalid OTP", http.StatusUnauthorized)
		return
	}

	// ✅ Update dispatch as verified
	_, err = dbPool.Exec(context.Background(),
		`UPDATE dispatches SET verified=TRUE WHERE id=$1`, dispatchID)
	if err != nil {
		http.Error(w, "Failed to update dispatch: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Mark trip as completed
	var tripID int
	err = dbPool.QueryRow(context.Background(),
		`UPDATE trips SET status='completed', last_updated=NOW()
		 WHERE dispatch_id=$1 RETURNING id`, dispatchID).Scan(&tripID)
	if err != nil {
		http.Error(w, "Failed to update trip: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Auto-create delivery record
	var deliveryID int
	var deliveryDate time.Time
	err = dbPool.QueryRow(context.Background(),
		`INSERT INTO deliveries (dispatch_id, trip_id, date)
		 VALUES ($1, $2, NOW())
		 RETURNING id, date`,
		dispatchID, tripID,
	).Scan(&deliveryID, &deliveryDate)
	if err != nil {
		http.Error(w, "Failed to create delivery: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Final response
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "OTP Verified ✅ Delivery completed",
		"dispatch": dispatchID,
		"trip":     tripID,
		"delivery": map[string]interface{}{
			"id":   deliveryID,
			"date": deliveryDate,
		},
	})
}

func RegisterDispatchRoutes(r *mux.Router) {
	r.HandleFunc("/dispatches", CreateDispatch).Methods("POST")
	r.HandleFunc("/dispatches", GetDispatches).Methods("GET")
	r.HandleFunc("/dispatches/{id}", GetDispatch).Methods("GET")
	r.HandleFunc("/dispatches/{id}", UpdateDispatch).Methods("PUT")
	r.HandleFunc("/dispatches/{id}", DeleteDispatch).Methods("DELETE")

	// OTP routes
	r.HandleFunc("/dispatches/{id}/send-otp", SendOTP).Methods("POST")
	r.HandleFunc("/dispatches/{id}/verify-otp", VerifyOTP).Methods("POST")
}
