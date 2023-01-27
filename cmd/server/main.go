// Auteurs: Jonathan Friedli, Lazar Pavicevic
// Labo 4 SDR

// Package main est le point d'entrée du programme permettant de démarrer le serveur.
// Le serveur a à disposition un fichier de configuration qui contient les adresses des serveurs ainsi que une liste d'adjacence représentant le graphe du réseau.
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

// main est la méthode d'entrée du programme
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

	_, ok := configuration.Servers[number]
	if !ok {
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
