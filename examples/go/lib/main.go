package main

import (
	"fmt"

	"github.com/cssbruno/gocep/pkg/cep"
)

func main() {
	result, address, err := cep.Search("06233903")
	fmt.Println(err)
	fmt.Println(result)  // json
	fmt.Println(address) // normalized address object
}
