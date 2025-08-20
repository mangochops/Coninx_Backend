package Client

import (
	"fmt"
	"time"
)

type delivery struct{
	recepient string
	condition string
	date time.Time
}

func Delivery (){
	fmt.Println("Goods delivered successfully")
}