package Admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type Register struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	var reg Register
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	// TODO: Save reg to database
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User " + reg.Email + " registered successfully!"))
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	// TODO: Authenticate user with database
	w.Write([]byte("User " + creds.Email + " logged in!"))
}

// RegisterAuthRoutes registers the auth endpoints to the router
func RegisterAuthRoutes(r *mux.Router) {
	r.HandleFunc("/signup", SignupHandler).Methods("POST")
	r.HandleFunc("/login", LoginHandler).Methods("POST")
}
