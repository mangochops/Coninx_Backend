package Admin

import (
	"context"
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
	ID        int           `json:"id"`
	Recepient string        `json:"recepient"`
	Phone     string        `json:"phone"`
	Location  string        `json:"location"`
	Driver    Driver.Driver `json:"driver"`
	Vehicle   Vehicle       `json:"vehicle"`
	Invoice   int           `json:"invoice"`
	Date      time.Time     `json:"date"`
	Verified  bool          `json:"verified"`
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

// ---------------- CRUD ----------------

func CreateDispatch(w http.ResponseWriter, r *http.Request) {
	var d Dispatch
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	err := dbPool.QueryRow(
		context.Background(),
		`INSERT INTO dispatches (recipient, phone, location, driver_id, vehicle_id, invoice, verified)
		 VALUES ($1,$2,$3,$4,$5,$6,FALSE)
		 RETURNING id, date, verified`,
		d.Recepient, d.Phone, d.Location, d.Driver.IDNumber, d.Vehicle.ID, d.Invoice,
	).Scan(&d.ID, &d.Date, &d.Verified)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(d)
}

func GetDispatches(w http.ResponseWriter, r *http.Request) {
	rows, err := dbPool.Query(context.Background(),
		`SELECT id, recipient, phone, location, invoice, date, verified FROM dispatches`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var dispatches []Dispatch
	for rows.Next() {
		var d Dispatch
		if err := rows.Scan(&d.ID, &d.Recepient, &d.Phone, &d.Location, &d.Invoice, &d.Date, &d.Verified); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		dispatches = append(dispatches, d)
	}

	json.NewEncoder(w).Encode(dispatches)
}

func GetDispatch(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	var d Dispatch
	err := dbPool.QueryRow(context.Background(),
		`SELECT id, recipient, phone, location, invoice, date, verified FROM dispatches WHERE id=$1`, id,
	).Scan(&d.ID, &d.Recepient, &d.Phone, &d.Location, &d.Invoice, &d.Date, &d.Verified)

	if err != nil {
		http.Error(w, "Dispatch not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(d)
}

func UpdateDispatch(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	var updated Dispatch
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	_, err := dbPool.Exec(context.Background(),
		`UPDATE dispatches SET recipient=$1, phone=$2, location=$3, invoice=$4 WHERE id=$5`,
		updated.Recepient, updated.Phone, updated.Location, updated.Invoice, id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updated.ID = id
	json.NewEncoder(w).Encode(updated)
}

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
	id, _ := strconv.Atoi(idStr)

	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var phone string
	err := dbPool.QueryRow(context.Background(),
		`SELECT phone FROM dispatches WHERE id=$1`, id).Scan(&phone)
	if err != nil {
		http.Error(w, "Dispatch not found", http.StatusNotFound)
		return
	}

	params := &openapi.CreateVerificationCheckParams{}
	params.SetTo(phone)
	params.SetCode(body.Code)

	resp, err := client.VerifyV2.CreateVerificationCheck(serviceSid, params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if *resp.Status == "approved" {
		_, err := dbPool.Exec(context.Background(),
			`UPDATE dispatches SET verified=TRUE WHERE id=$1`, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{
			"message": "OTP Verified, delivery confirmed ✅",
		})
		return
	}

	http.Error(w, "Invalid OTP", http.StatusUnauthorized)
}


// RegisterDispatchRoutes registers the dispatch endpoints to the router
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
	


