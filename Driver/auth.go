package Driver

import (
	"encoding/json"
	
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Driver struct {
	FirstName           string `json:"firstName"`
	LastName            string `json:"lastName"`
	PhoneNumber         int    `json:"phoneNumber"`
	Password            string `json:"password"`
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
	// TODO: Save driver to database
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Driver " + d.FirstName + " registered successfully!"))
}

// Optionally, add a login handler for drivers
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	var creds struct {
		PhoneNumber int    `json:"phoneNumber"`
		Password    string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	// TODO: Authenticate driver with database
	w.Write([]byte("Driver " + strconv.Itoa(creds.PhoneNumber) + " logged in!"))
	// Optionally, remove the following line if not needed
	// w.Write([]byte("Driver " + string(creds.PhoneNumber) + " logged in!"))
}

// RegisterDriverRoutes registers the driver endpoints to the router
func RegisterDriverRoutes(r *mux.Router) {
	r.HandleFunc("/driver/register", RegisterHandler).Methods("POST")
	r.HandleFunc("/driver/login", LoginHandler).Methods("POST")
}


