package driver

import (
	"fmt"
)

type Driver struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func Register(){
	fmt.Println("Driver created successfully")
}