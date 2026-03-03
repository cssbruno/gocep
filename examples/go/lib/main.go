package main

import (
	"fmt"
	"time"

	"github.com/cssbruno/gocep/pkg/cep"
)

func main() {
	opts := cep.GetOptions()
	opts.SearchTimeout = 5 * time.Second
	cep.SetOptions(opts)

	result, address, err := cep.Search("06233903")
	fmt.Println(err)
	fmt.Println(result)  // json
	fmt.Println(address) // normalized address object
}
