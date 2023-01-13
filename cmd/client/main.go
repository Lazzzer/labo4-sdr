package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) != 1 {
		log.Fatal("Usage: go run main.go")
	}
	/*cl := client.Client()
	cl.Run()*/
}
