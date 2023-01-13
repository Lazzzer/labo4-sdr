package main

import (
	"log"
	"os"
	"strconv"
)

func main() {
	errorMessage := "Usage: go run main.go <server number>"
	if len(os.Args) != 2 {
		log.Fatal(errorMessage)
	}

	number, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal(errorMessage)
	} else if number < 0 || number > 5 { // todo ne pas hardcoder le 5
		log.Fatal("Number must be between 0 and 5")
	}

	/*
		server := server.Server(number)
		server.Run()
	*/
}
