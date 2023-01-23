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

type Client struct {
	Servers map[int]string
}

var exitChan = make(chan os.Signal, 1) // Chan qui gère le ctrl + c

func (c *Client) Run() {
	signal.Notify(exitChan, syscall.SIGINT)

	go func() {
		<-exitChan
		fmt.Println("Bye, have a great time.")
		os.Exit(0)
	}()

	fmt.Println("SDR - Labo 4 - Client")
	reader := bufio.NewReader(os.Stdin)
	for {
		displayPrompt()
		input, err := reader.ReadString('\n')
		if err != nil {
			shared.Log(types.ERROR, err.Error())
			continue
		}
		waitResponse, command, addresses, err := processInput(input, c)
		if err != nil {
			shared.Log(types.ERROR, err.Error())
			continue
		}

		for _, servAddr := range addresses {
			c.sendCommand(command, servAddr, waitResponse)
		}
	}
}

func displayPrompt() {
	fmt.Println("\n Available commands:")
	fmt.Println(" - wave <word>")
	fmt.Println(" - probe <word> <serverToAsk>")
	fmt.Println(" - ask <serverToAsk>")
	fmt.Println(" - quit")
	fmt.Println("Enter a command to send:")
}

func processInput(input string, c *Client) (bool, string, []string, error) {
	args := strings.Fields(input)
	length := len(args)

	// String vide
	if length == 0 {
		return false, "", nil, fmt.Errorf("empty input")
	}

	// Quit
	if args[0] == string(types.Quit) {
		exitChan <- syscall.SIGINT
	}
	var command types.Command
	var addresses []string
	waitResponse := false

	switch args[0] {
	case string(types.WaveCount):
		if length == 2 {
			command.Type = types.WaveCount
			command.Text = args[1]
			for _, address := range c.Servers {
				addresses = append(addresses, address)
			}
		} else {
			return false, "", nil, fmt.Errorf("invalid wave command")
		}
	case string(types.ProbeCount):
		if length == 3 {
			value, err := strconv.Atoi(args[2])
			if err != nil || value < 1 || value > len(c.Servers) {
				return false, "", nil, fmt.Errorf("invalid server number")
			}
			command.Type = types.ProbeCount
			command.Text = args[1]
			addresses = append(addresses, c.Servers[value])
			waitResponse = true
		} else {
			return false, "", nil, fmt.Errorf("invalid probe command")
		}
	case string(types.Ask):
		if length == 2 {
			value, err := strconv.Atoi(args[1])
			if err != nil || value < 1 || value > len(c.Servers) {
				return false, "", nil, fmt.Errorf("invalid server number")
			}
			command.Type = types.Ask
			command.Text = ""
			addresses = append(addresses, c.Servers[value])
			waitResponse = true
		} else {
			return false, "", nil, fmt.Errorf("invalid ask command")
		}
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
// Elle a également la possibilité de timeout si le serveur ne répond pas.
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
			// TODO: Vérifier le comportement
			fmt.Println(shared.RED + "Server @" + udpAddr.String() + " is unreachable" + shared.RESET)
			return
		}

		fmt.Println(shared.GREEN + "From Server @" + servAddr.String() + "\n" + shared.RESET + string(buffer[0:n]))
	}
}
