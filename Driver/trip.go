package Driver

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Coordinates and Trip structs
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
	tripsMu sync.Mutex

	clients   = make(map[*websocket.Conn]int) // track subscribed driverId
	clientsMu sync.Mutex
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

// WebSocket handler
func TripWSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WS upgrade error:", err)
		return
	}
	defer conn.Close()

	// Register client with no subscription initially
	clientsMu.Lock()
	clients[conn] = 0
	clientsMu.Unlock()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msgData struct {
			Type      string  `json:"type"`
			DriverID  int     `json:"driverId"`
			TripID    int     `json:"tripId,omitempty"`
			Latitude  float64 `json:"latitude,omitempty"`
			Longitude float64 `json:"longitude,omitempty"`
		}

		if err := json.Unmarshal(msg, &msgData); err != nil {
			log.Println("WS unmarshal error:", err)
			continue
		}

		switch msgData.Type {
		case "subscribe":
			// Subscribe client to a specific driver
			clientsMu.Lock()
			clients[conn] = msgData.DriverID
			clientsMu.Unlock()

		case "update_location":
			// Update trip coordinates
			tripsMu.Lock()
			trip, ok := trips[msgData.TripID]
			if ok {
				trip.Coordinates = Coordinates{Latitude: msgData.Latitude, Longitude: msgData.Longitude}
				trips[msgData.TripID] = trip
			}
			tripsMu.Unlock()

			if ok {
				broadcastDriverUpdate(trip)
			}
		}
	}

	// Remove client when disconnected
	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()
}

// Broadcast only to clients subscribed to this driver
func broadcastDriverUpdate(trip Trip) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	data, _ := json.Marshal(trip)
	for client, driverID := range clients {
		if driverID == trip.DriverID {
			client.WriteMessage(websocket.TextMessage, data)
		}
	}
}

// Register trip routes (including WS)
func RegisterTripRoutes(r *mux.Router) {
	r.HandleFunc("/ws/trips", TripWSHandler)
}
