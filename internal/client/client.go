// Auteurs: Jonathan Friedli, Lazar Pavicevic
// Labo 4 SDR

// Package client propose un client UDP envoyant des commandes sous forme de string json à des serveurs du réseau.
//
// Le client parse l'entrée de l'utilisateur et envoie la commande correspondante au serveur.
package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/Lazzzer/labo4-sdr/internal/shared"
	"github.com/Lazzzer/labo4-sdr/internal/shared/types"
)

// Client représente un client UDP connecté à un réseau de serveurs capable d'envoyer des commandes de traitement de texte.
type Client struct {
	Servers map[int]string // Map des serveurs du réseau, avec comme clé le numéro du serveur et comme valeur l'adresse du serveur
}

var exitChan = make(chan os.Signal, 1) // Chan qui gère le CTRL+C

// Run est la méthode principale du client. Elle gère l'entrée de l'utilisateur et envoie les commandes aux serveurs.
func (c *Client) Run() {
	signal.Notify(exitChan, syscall.SIGINT)

	go func() {
		<-exitChan
		fmt.Println("\nBye, have a great time.")
		os.Exit(0)
	}()

	fmt.Println(shared.BOLD + "SDR - Labo 4 - Client" + shared.RESET)
	reader := bufio.NewReader(os.Stdin)
	for {
		displayPrompt()
		input, err := reader.ReadString('\n')
		if err != nil {
			shared.Log(types.ERROR, err.Error())
			continue
		}
		waitResponse, command, addresses, err := c.processInput(input)
		if err != nil {
			fmt.Println(shared.RED + "\nERROR: " + err.Error() + shared.RESET)
			continue
		}

		for _, servAddr := range addresses {
			c.sendCommand(command, servAddr, waitResponse)
		}
	}
}

// processInput parse l'entrée de l'utilisateur et retourne un tuple contenant :
// - un booléen indiquant si le client doit attendre une réponse du serveur
// - la commande à envoyer au serveur sous forme de string json
// - l'adresse du serveur auquel envoyer la commande
// - une erreur si l'entrée est invalide
func (c *Client) processInput(input string) (bool, string, []string, error) {
	args := strings.Fields(input)
	length := len(args)

	// String vide
	if length == 0 {
		return false, "", nil, fmt.Errorf("empty input")
	}

	var command types.Command

	var addresses []string
	waitResponse := false

	switch args[0] {
	case string(types.WaveCount):

		if length < 2 {
			return false, "", nil, fmt.Errorf("invalid wave command")
		}

		command.Text = strings.Join(args[1:], " ")

		command.Type = types.WaveCount
		for _, address := range c.Servers {
			addresses = append(addresses, address)
		}
	case string(types.ProbeCount):
		if length < 3 {
			return false, "", nil, fmt.Errorf("invalid probe command")
		}

		value, err := strconv.Atoi(args[1])
		if err != nil {
			return false, "", nil, fmt.Errorf("invalid server number")
		}
		if _, ok := c.Servers[value]; !ok {
			return false, "", nil, fmt.Errorf("invalid server number")
		}

		command.Text = strings.Join(args[2:], " ")

		command.Type = types.ProbeCount
		addresses = append(addresses, c.Servers[value])
		waitResponse = true
	case string(types.Ask):
		if length != 2 {
			return false, "", nil, fmt.Errorf("invalid ask command")
		}
		value, err := strconv.Atoi(args[1])
		if err != nil {
			return false, "", nil, fmt.Errorf("invalid server number")
		}
		if _, ok := c.Servers[value]; !ok {
			return false, "", nil, fmt.Errorf("invalid server number")
		}

		command.Type = types.Ask
		command.Text = ""
		addresses = append(addresses, c.Servers[value])
		waitResponse = true
	case string(types.Quit):
		fmt.Println("\nBye, have a great time.")
		os.Exit(0)
	default:
		return false, "", nil, fmt.Errorf("unknown command")
	}

	if jsonCommand, err := json.Marshal(command); err == nil {
		return waitResponse, string(jsonCommand), addresses, nil
	} else {
		return false, "", nil, fmt.Errorf("error while marshalling command")
	}
}

// sendCommand envoie une commande au serveur spécifié. Elle s'occupe de la connexion UDP et de la fermeture de celle-ci.
func (c *Client) sendCommand(command string, address string, waitResponse bool) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatal(err)
	}

	connection, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		shared.Log(types.ERROR, err.Error())
		return
	}
	defer func(connection *net.UDPConn) {
		err := connection.Close()
		if err != nil {
			shared.Log(types.ERROR, err.Error())
		}
	}(connection)

	_, err = connection.Write([]byte(command + "\n"))
	if err != nil {
		shared.Log(types.ERROR, err.Error())
		return
	}

	if waitResponse {
		buffer := make([]byte, 1024)
		n, servAddr, err := connection.ReadFromUDP(buffer)

		if err != nil {
			fmt.Println(shared.RED + "\nServer @" + udpAddr.String() + " is unreachable" + shared.RESET)
			return
		}

		fmt.Println(shared.GREEN + "\nFrom Server @" + servAddr.String() + "\n" + shared.RESET + string(buffer[0:n]))
	}
}

// displayPrompt affiche les commandes disponibles pour l'utilisateur.
func displayPrompt() {
	fmt.Println("\nAvailable commands:")
	fmt.Println(shared.YELLOW + " - wave <text>")
	fmt.Println(" - probe <server number> <text>")
	fmt.Println(" - ask <server number>")
	fmt.Println(" - quit" + shared.RESET)
	fmt.Println(shared.BOLD + "\nEnter a command to send:" + shared.RESET)
}
