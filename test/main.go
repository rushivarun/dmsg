package main

import (
	"encoding/json"
	"fmt"
)

type Employee struct {
	Name   string `json:"empname"`
	Number int    `json:"empid"`
}

func main() {
	emp := &Employee{Name: "Rocky", Number: 5454}
	e, err := json.Marshal(emp)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(e))
}
