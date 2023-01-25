// Auteurs: Jonathan Friedli, Lazar Pavicevic
// Labo 4 SDR

// Package main est le point d'entrée du programme permettant de démarrer le client.
// Le client a à disposition un fichier de configuration qui contient les adresses des serveurs.
package main

import (
	_ "embed"
	"log"
	"os"

	"github.com/Lazzzer/labo4-sdr/internal/client"
	"github.com/Lazzzer/labo4-sdr/internal/shared"
	"github.com/Lazzzer/labo4-sdr/internal/shared/types"
)

//go:embed config.json
var config string

// main est la méthode d'entrée du programme
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
