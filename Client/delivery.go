package Client

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
	"github.com/mangochops/coninx_backend/Admin"
)

type Delivery struct {
	ID           int       `json:"id"`
	DispatchNumber  Admin.Dispatch `json:"dispatchNumber"`
	PhoneNumber  int       `json:"phoneNumber"`
	Code int `json:"code"`
	Date         time.Time `json:"date"`
}

func CreateRecepient(w http.ResponseWriter, r *http.Request){
	
}

func GetRecepient(w http.ResponseWriter, r *http.Request){

}

func DeleteRecepient(w http.ResponseWriter, r *http.Request){

}
func DeliverGoods() {
	fmt.Println("Goods delivered successfully")
}
