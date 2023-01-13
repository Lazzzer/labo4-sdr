package main

import (
	_ "embed"
	"github.com/Lazzzer/labo4-sdr/internal/client"
	"github.com/Lazzzer/labo4-sdr/internal/shared"
	"github.com/Lazzzer/labo4-sdr/internal/shared/types"
	"log"
	"os"
)

//go:embed config.json
var config string

func main() {
	if len(os.Args) != 1 {
		log.Fatal("Usage: go run main.go")
	}

	configuration, err := shared.Parse[types.Config](config)
	if err != nil {
		log.Fatal(err)
	}

	cl := client.Client{
		Servers: configuration.Servers,
	}
	cl.Run()
}
