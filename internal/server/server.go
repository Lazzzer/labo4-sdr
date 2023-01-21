package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/Lazzzer/labo4-sdr/internal/shared"
	"github.com/Lazzzer/labo4-sdr/internal/shared/types"
)

var waveMessageChans = make(map[int](chan types.Message)) // Map de channels qui gère les messages wave pour chaque processus

// TODO : Refactor to have subservers for each algo
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

// TODO : Refactor to launch for wave and probe commands
func (s *Server) Init(adjacencyList *map[int][]int) {

	// Initialisation de la map des voisins et des voisins actifs
	s.Neighbors = make(map[int]types.Server)
	s.ActiveNeighbors = make(map[int]types.Server)

	for i := 0; i < len((*adjacencyList)[s.Number]); i++ {
		s.Neighbors[(*adjacencyList)[s.Number][i]] = s.Servers[(*adjacencyList)[s.Number][i]]
		s.ActiveNeighbors[(*adjacencyList)[s.Number][i]] = s.Servers[(*adjacencyList)[s.Number][i]]
		waveMessageChans[(*adjacencyList)[s.Number][i]] = make(chan types.Message, 1)
	}

	// Initialisation du nombre de voisins
	s.NbNeighbors = len(s.Neighbors)

	// Initialisation de la map de la topologie
	s.Topology = make(map[string]int)
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

func (s *Server) countLetterOccurrences(text string) {
	shared.Log(types.INFO, "Counting occurrences of letter "+s.Letter+" in "+text+"...")
	s.Topology[s.Letter] = strings.Count(strings.ToUpper(text), s.Letter)
	shared.Log(types.INFO, "Letter "+s.Letter+" found "+strconv.Itoa(s.Topology[s.Letter])+" time(s) in "+text)
}

// handleMessage gère les messages reçus des autres serveurs en UDP.
func (s *Server) handleMessage(connection *net.UDPConn, addr *net.UDPAddr, messageStr string) error {
	message, err := shared.Parse[types.Message](messageStr)

	// TODO: Refactor this
	if err != nil || message.Number == 0 {
		return fmt.Errorf("invalid message type")
	}

	waveMessageChans[message.Number] <- *message
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

// TODO : Refactor to reset topology and neighbors then to fit the second algo later
// handleWaveCount
func (s *Server) handleWaveCount(text string) {
	s.countLetterOccurrences(text)

	for len(s.Topology) < s.NbProcesses {

		// Tri des clés des maps pour les envoyer dans l'ordre
		var keysNeighbors []int
		var keysActiveNeighbors []int
		for key := range s.Neighbors {
			keysNeighbors = append(keysNeighbors, key)
		}
		for key := range s.ActiveNeighbors {
			keysActiveNeighbors = append(keysActiveNeighbors, key)
		}
		sort.Ints(keysNeighbors)
		sort.Ints(keysActiveNeighbors)

		message := types.Message{
			Topology: s.Topology,
			Number:   s.Number,
			Active:   true,
		}
		for _, key := range keysNeighbors {
			s.sendWaveMessage(message, s.Neighbors[key])
		}
		for _, key := range keysActiveNeighbors {
			shared.Log(types.WAVE, "Waiting for P"+strconv.Itoa(key)+" to send its topology...")
			message := <-waveMessageChans[key]
			for letter, count := range message.Topology {
				s.Topology[letter] = count
			}
			if !message.Active {
				delete(s.ActiveNeighbors, message.Number)
			}
		}
	}
	message := types.Message{
		Topology: s.Topology,
		Number:   s.Number,
		Active:   false,
	}

	var keysActiveNeighbors []int
	for key := range s.ActiveNeighbors {
		keysActiveNeighbors = append(keysActiveNeighbors, key)
	}
	sort.Ints(keysActiveNeighbors)

	shared.Log(types.WAVE, "Sending final topology to active neighbors...")
	// Envoi de la topologie aux voisins actifs
	for _, key := range keysActiveNeighbors {
		s.sendWaveMessage(message, s.ActiveNeighbors[key])
		shared.Log(types.WAVE, "Sent final topology to P"+strconv.Itoa(key))
	}

	shared.Log(types.WAVE, "Purging messages from active neighbors...")
	// Purge des derniers messages reçus
	for _, key := range keysActiveNeighbors {
		<-waveMessageChans[key]
		shared.Log(types.WAVE, "Purged message from P"+strconv.Itoa(key))
	}
	// TODO: Format output of topology to be readable in the console
	shared.Log(types.WAVE, "Text "+text+" has been processed.")
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

// TODO: Put logs maybe ?
func (s *Server) sendWaveMessage(message types.Message, neighbor types.Server) error {
	messageJson, err := json.Marshal(message)
	if err != nil {
		shared.Log(types.ERROR, err.Error())
		return err
	}

	destUdpAddr, err := net.ResolveUDPAddr("udp", neighbor.Address)
	if err != nil {
		return err
	}
	connection, err := net.DialUDP("udp", nil, destUdpAddr)
	if err != nil {
		return err
	}
	_, err = connection.Write(messageJson)
	if err != nil {
		return err
	}
	return nil
}
