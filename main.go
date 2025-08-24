package main

import(
	"fmt"
	"net/http"
	"log"
	"github.com/mangochops/coninx_backend/Admin"
)

func main() {
	http.HandleFunc("/signup", Admin.SignupHandler)
	http.HandleFunc("/login", Admin.LoginHandler)
	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}