package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rzeradev/multithreading/pkg/cepservice"
)

func main() {
	cep := "70150900"
	if len(os.Args) > 1 {
		cep = os.Args[1]
	}
	timeout := 1 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result, err := cepservice.GetFasterAPIResult(ctx, cep)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("API:", result.API)
	fmt.Printf("Address: %+v\n", result.Address)
}
