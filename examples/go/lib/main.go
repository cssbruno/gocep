package main

import (
	"fmt"

	"github.com/cssbruno/gocep/pkg/cep"
)

func main() {
	result, wecep, err := cep.Search("06233903")
	fmt.Println(err)
	fmt.Println(result) // json
	fmt.Println(wecep)  // WeCep object
}
