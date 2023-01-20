package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/Lazzzer/labo4-sdr/internal/shared"
	"github.com/Lazzzer/labo4-sdr/internal/shared/types"
)

type Server struct {
	Number          int                  `json:"number"`           // Numéro du processus
	NbProcesses     int                  `json:"nb_processes"`     // Nombre total de processus
	Letter          string               `json:"letter"`           // Lettre à compter
	Address         string               `json:"address"`          // Adresse du serveur du processus
	Servers         map[int]types.Server `json:"servers"`          // Map des serveurs disponibles
	Neighbors       map[int]types.Server `json:"neighbors"`        // Map des serveurs voisins
	ActiveNeighbors map[int]types.Server `json:"active_neighbors"` // Map des serveurs voisins actifs
	NbNeighbors     int                  `json:"nb_neighbors"`     // Nombre de processus voisins
	Topology        map[string]int       `json:"topology"`         // Map de la topologie qui contient le compteur de chaque lettre
}

func (s *Server) Init(adjacencyList *map[int][]int) {

	// Initialisation de la map des voisins et des voisins actifs
	for i := 0; i < len((*adjacencyList)[s.Number]); i++ {
		s.Neighbors[(*adjacencyList)[s.Number][i]] = s.Servers[(*adjacencyList)[s.Number][i]]
		s.ActiveNeighbors[(*adjacencyList)[s.Number][i]] = s.Servers[(*adjacencyList)[s.Number][i]]
	}

	// Initialisation du nombre de voisins
	s.NbNeighbors = len(s.Neighbors)
}

func (s *Server) Run() {
	connection := s.startListening()
	defer connection.Close()

	shared.Log(types.INFO, shared.GREEN+"Server #"+strconv.Itoa(s.Number)+" as Process P"+strconv.Itoa(s.Number)+" listening on "+s.Address+shared.RESET)

	for {
		s.handleCommunications(connection)
	}
}

// startListening initialise la connexion UDP du serveur et écoute les connexions entrantes.
func (s *Server) startListening() *net.UDPConn {
	udpAddr, err := net.ResolveUDPAddr("udp4", s.Address)
	if err != nil {
		log.Fatal(err)
	}

	connection, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		log.Fatal(err)
	}

	return connection
}

// handleCommunications gère les communications du serveur.
// La méthode écoute les messages entre serveurs ainsi que les commandes des clients.
func (s *Server) handleCommunications(connection *net.UDPConn) {
	buffer := make([]byte, 1024)
	for {
		n, addr, err := connection.ReadFromUDP(buffer)
		if err != nil {
			shared.Log(types.ERROR, err.Error())
			continue
		}
		communication := string(buffer[0:n])
		err = s.handleMessage(connection, addr, communication)
		if err != nil {
			go func() {
				// Traitement d'une commande si le message n'est pas valide
				response, err := s.handleCommand(communication)
				if err != nil {
					shared.Log(types.ERROR, err.Error())
				}
				// Envoi de la réponse à l'adresse du client si elle existe
				if response != "" {
					_, err = connection.WriteToUDP([]byte(response), addr)
					if err != nil {
						shared.Log(types.ERROR, err.Error())
					}
				}
			}()
		}
	}
}

func (s *Server) countLetterOccurences(text string) int {
	shared.Log(types.INFO, "counting letter "+s.Letter+" occurrences...")
	for _, letter := range strings.ToLower(text) {
		if string(letter) == s.Letter {
			s.Topology[s.Letter]++
		}
	}
	return s.Topology[s.Letter]
}

// handleMessage gère les messages reçus des autres serveurs en UDP.
func (s *Server) handleMessage(connection *net.UDPConn, addr *net.UDPAddr, messageStr string) error {
	message, err := shared.Parse[types.Message](messageStr)

	// TODO: Refactor this
	if err != nil || message.Number == 0 {
		return fmt.Errorf("invalid message type")
	}
	return nil
}

// handleCommand gère les commandes reçues des clients UDP.
func (s *Server) handleCommand(commandStr string) (string, error) {
	command, err := shared.Parse[types.Command](commandStr)
	if err != nil || command.Type == "" {
		return "", fmt.Errorf("invalid command")
	}

	shared.Log(types.COMMAND, "GOT => Type: "+string(command.Type)+" Text: "+command.Text)

	switch command.Type {
	case types.WaveCount:
		go s.handleWaveCount(command.Text)
	case types.ProbeCount:
		go s.handleProbeCount(command.Text)
	case types.Ask:
		return s.handleAsk(command.Text), nil
	}
	return "", nil
}

// handleWaveCount
func (s *Server) handleWaveCount(text string) string {
	return ""
}

// handleProbeCount
func (s *Server) handleProbeCount(text string) string {
	// TODO
	return ""
}

func (s *Server) handleAsk(text string) string {
	// TODO
	return ""
}
