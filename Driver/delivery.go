package Driver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type Delivery struct {
	ID           int       `json:"id"`
	DeliveryNote string    `json:"deliveryNote"` // file path to the uploaded image
	Recipient    string    `json:"recipient"`
	Condition    string    `json:"condition"`
	Date         time.Time `json:"date"`
}

var (
	deliveries   = make(map[int]Delivery)
	// nextID is defined in trip.go and shared across the package
	deliveriesMu sync.Mutex
)

// Handle delivery creation with image upload
func CreateDeliveryHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		http.Error(w, "Could not parse multipart form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("deliveryNote")
	if err != nil {
		http.Error(w, "Could not get deliveryNote file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save the uploaded file
	uploadDir := "./uploads"
	os.MkdirAll(uploadDir, os.ModePerm)
	filePath := filepath.Join(uploadDir, strconv.Itoa(nextID)+"_"+handler.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Could not save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Could not write file", http.StatusInternalServerError)
		return
	}

	recipient := r.FormValue("recipient")
	condition := r.FormValue("condition")

	deliveriesMu.Lock()
	d := Delivery{
		ID:           nextID,
		DeliveryNote: filePath,
		Recipient:    recipient,
		Condition:    condition,
		Date:         time.Now(),
	}
	deliveries[nextID] = d
	nextID++
	deliveriesMu.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(d)
}

// Get a delivery by ID
func GetDeliveryHandler(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid delivery ID", http.StatusBadRequest)
		return
	}
	deliveriesMu.Lock()
	d, ok := deliveries[id]
	deliveriesMu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	json.NewEncoder(w).Encode(d)
}

// List all deliveries
func ListDeliveriesHandler(w http.ResponseWriter, r *http.Request) {
	deliveriesMu.Lock()
	defer deliveriesMu.Unlock()
	var list []Delivery
	for _, d := range deliveries {
		list = append(list, d)
	}
	json.NewEncoder(w).Encode(list)
}

// RegisterDeliveryRoutes registers all delivery endpoints to the router
func RegisterDeliveryRoutes(r *mux.Router) {
	r.HandleFunc("/delivery", CreateDeliveryHandler).Methods("POST")
	r.HandleFunc("/delivery/{id}", GetDeliveryHandler).Methods("GET")
	r.HandleFunc("/deliveries", ListDeliveriesHandler).Methods("GET")
}

func DeliverGoods() {
	fmt.Println("Goods delivered successfully")
}
