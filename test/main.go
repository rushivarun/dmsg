package main

import (
	"fmt"
)

type mad struct {
	mes []string
}

var m mad

func main() {
	m.mes = append(m.mes, "one")
	m.mes = append(m.mes, "two")
	m.mes = append(m.mes, "three")
	m.mes = append(m.mes, "four")
	m.mes = append(m.mes, "fove")

	for idx, val := range m.mes {
		fmt.Println(idx, ",", val)
		if val == "four" {
			fmt.Println(m.mes[idx])
		}
	}
}
