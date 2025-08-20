package Admin

import (
	"fmt"
	"time"
)

type dispatch struct{
	ID int
	recepient  string
	location  string
	// items 
	quantity  int
	price int
	date time.Time

}

func Dispatch(){
	fmt.Println("Dispatch created")
}