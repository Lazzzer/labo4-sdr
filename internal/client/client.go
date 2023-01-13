package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Lazzzer/labo4-sdr/internal/shared"
	"github.com/Lazzzer/labo4-sdr/internal/shared/types"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
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
		command, serverAddress, err := processInput(input, c)
		if err != nil {
			shared.Log(types.ERROR, err.Error())
			continue
		}
		println("Command is : " + command + " and server address is : " + serverAddress)
	}
}

func displayPrompt() {
	fmt.Println("\n Available commands:")
	fmt.Println(" - count <word>")
	fmt.Println(" - count <word> <serverToAsk>")
	fmt.Println(" - ask <serverToAsk>")
	fmt.Println(" - quit")
	fmt.Println("Enter a command to send:")
}

func processInput(input string, c *Client) (string, string, error) {
	args := strings.Fields(input)
	length := len(args)

	// String vide
	if length == 0 {
		return "", "", fmt.Errorf("empty input")
	}

	// Quit
	if args[0] == string(types.Quit) {
		exitChan <- syscall.SIGINT
	}
	var command types.Command

	switch args[0] {
	case string(types.Count):
		if length == 2 {
			command.Type = types.Count
			command.Text = args[1]
			value := 1 //TODO: c'est de la merde ça
			command.Server = &value
		} else if length == 3 {
			value, err := strconv.Atoi(args[2])
			if err != nil || value < 1 || value > len(c.Servers) {
				return "", "", fmt.Errorf("invalid server number")
			}
			command.Type = types.Count
			command.Text = args[1]
			command.Server = &value
		} else {
			return "", "", fmt.Errorf("invalid command")
		}
		break
	case string(types.Ask):
		if length != 2 {
			return "", "", fmt.Errorf("invalid command")
		}
		value, err := strconv.Atoi(args[1])
		if err != nil || value < 1 || value > len(c.Servers) {
			return "", "", fmt.Errorf("invalid server number")
		}
		command.Type = types.Ask
		command.Text = ""
		command.Server = &value
		break
	default:
		return "", "", fmt.Errorf("unknown command")
	}

	/*if length == 2 {
		if args[0] == string(types.Count) {
			command.Type = types.Count
			command.Text = args[1]
			value := 1 //TODO: c'est de la merde ça
			command.Server = &value
		} else {
			return "", "", fmt.Errorf("invalid command")
		}
	} else if length == 3 {
		value, err := strconv.Atoi(args[2])
		if err != nil || value < 1 || value > len(c.Servers) {
			return "", "", fmt.Errorf("invalid server number")
		}
		if args[0] == string(types.Count) {
			command.Type = types.Count
			command.Text = args[1]
			command.Server = &value
		} else if args[0] == string(types.Ask) {
			command.Type = types.Ask
			command.Text = ""
			command.Server = &value
		} else {
			return "", "", fmt.Errorf("invalid command")
		}
	} else {
		return "", "", fmt.Errorf("invalid command")
	}*/

	if jsonCommand, err := json.Marshal(command); err == nil {
		return string(jsonCommand), c.Servers[*command.Server], nil
	} else {
		return "", "", fmt.Errorf("error while marshalling command")
	}
}
