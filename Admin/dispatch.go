package Admin

import (
	"encoding/json"

	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/mangochops/coninx_backend/Driver"
)

type Dispatch struct {
	ID        int       `json:"id"`
	Recepient string    `json:"recepient"`
	Location  string    `json:"location"`
	Driver    Driver.Driver `json:"driver"`
	Vehicle   Vehicle   `json:"vehicle"`
	Invoice   int       `json:"invoice"`
	Date      time.Time `json:"date"`
}

var dispatches []Dispatch
var nextID = 1

func CreateDispatch(w http.ResponseWriter, r *http.Request) {
	var d Dispatch
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	d.ID = nextID
	nextID++
	d.Date = time.Now()
	dispatches = append(dispatches, d)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(d)
}

func GetDispatches(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(dispatches)
}

func GetDispatch(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	for _, d := range dispatches {
		if d.ID == id {
			json.NewEncoder(w).Encode(d)
			return
		}
	}
	http.NotFound(w, r)
}

func UpdateDispatch(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	var updated Dispatch
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	for i, d := range dispatches {
		if d.ID == id {
			updated.ID = id
			updated.Date = d.Date
			dispatches[i] = updated
			json.NewEncoder(w).Encode(updated)
			return
		}
	}
	http.NotFound(w, r)
}

func DeleteDispatch(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	for i, d := range dispatches {
		if d.ID == id {
			dispatches = append(dispatches[:i], dispatches[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.NotFound(w, r)
}

// RegisterDispatchRoutes registers all dispatch endpoints to the router
func RegisterDispatchRoutes(r *mux.Router) {
	r.HandleFunc("/dispatch", CreateDispatch).Methods("POST")
	r.HandleFunc("/dispatch", GetDispatches).Methods("GET")
	r.HandleFunc("/dispatch/{id}", GetDispatch).Methods("GET")
	r.HandleFunc("/dispatch/{id}", UpdateDispatch).Methods("PUT")
	r.HandleFunc("/dispatch/{id}", DeleteDispatch).Methods("DELETE")
}


