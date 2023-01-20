package main

import (
	_ "embed"
	"flag"
	"log"
	"strconv"

	"github.com/Lazzzer/labo4-sdr/internal/server"
	"github.com/Lazzzer/labo4-sdr/internal/shared"
	"github.com/Lazzzer/labo4-sdr/internal/shared/types"
)

//go:embed config.json
var config string

func main() {
	flag.Parse()
	if flag.Arg(0) == "" {
		log.Fatal("Invalid argument, usage: <server number>")
	}

	number, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		log.Fatal("Invalid argument, usage: <server number>")
	}

	configuration, err := shared.Parse[types.ServerConfig](config)
	if err != nil {
		log.Fatal(err)
	}

	if number > len(configuration.Servers) || number < 0 {
		log.Fatal("Invalid server number")
	}

	server := server.Server{
		Number:      number,
		NbProcesses: len(configuration.Servers),
		Letter:      configuration.Servers[number].Letter,
		Address:     configuration.Servers[number].Address,
		Servers:     configuration.Servers,
	}

	server.Init(&configuration.AdjacencyList)
	server.Run()
}
