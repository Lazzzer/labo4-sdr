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
var textProcessedChan = make(chan bool, 1)                // Channel qui gère la fin du traitement du texte

// TODO : Refactor to have subservers for each algo
type Server struct {
	Number          int                  `json:"number"`           // Numéro du processus
	NbProcesses     int                  `json:"nb_processes"`     // Nombre total de processus
	Letter          string               `json:"letter"`           // Lettre à compter
	Address         string               `json:"address"`          // Adresse du serveur du processus
	Servers         map[int]types.Server `json:"servers"`          // Map des serveurs disponibles
	Parent          int                  `json:"parent"`           // Numéro du processus parent
	Neighbors       map[int]types.Server `json:"neighbors"`        // Map des serveurs voisins
	ActiveNeighbors map[int]types.Server `json:"active_neighbors"` // Map des serveurs voisins actifs
	NbNeighbors     int                  `json:"nb_neighbors"`     // Nombre de processus voisins
	Counts          map[string]int       `json:"counts"`           // Map qui contient le compteur de chaque lettre gérée par les processus
	Text            string               `json:"text"`             // Texte dont il faut compter l'occurrence des lettres
	TextProcessed   bool                 `json:"text_processed"`   // Booléen qui indique si le texte a été traité
}

// TODO : Refactor to launch for wave and probe commands
func (s *Server) Init(adjacencyList *map[int][]int) {

	// Initialisation du texte et de la variable isProcessing
	s.Text = ""
	textProcessedChan <- false

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

	// Initialisation de la map de compteurs
	s.Counts = make(map[string]int)
}

func (s *Server) Run() {
	connection := s.startListening()
	defer connection.Close()

	shared.Log(types.INFO, shared.GREEN+"Server #"+strconv.Itoa(s.Number)+" as Process P"+strconv.Itoa(s.Number)+" listening on "+s.Address+shared.RESET)

	s.handleCommunications(connection)
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
	s.Counts[s.Letter] = strings.Count(strings.ToUpper(text), s.Letter)
	shared.Log(types.INFO, "Letter "+s.Letter+" found "+strconv.Itoa(s.Counts[s.Letter])+" time(s) in "+text)
}

// handleMessage gère les messages reçus des autres serveurs en UDP.
func (s *Server) handleMessage(connection *net.UDPConn, addr *net.UDPAddr, messageStr string) error {
	message, err := shared.Parse[types.Message](messageStr)

	// TODO: Refactor this
	if err != nil || message.Number == 0 {
		return fmt.Errorf("invalid message type")
	}

	go func() {
		waveMessageChans[message.Number] <- *message
	}()

	return nil
}

// handleCommand gère les commandes reçues des clients UDP.
func (s *Server) handleCommand(commandStr string) (string, error) {
	command, err := shared.Parse[types.Command](commandStr)
	if err != nil || command.Type == "" {
		return "", fmt.Errorf("invalid command")
	}

	textToLog := ""
	if command.Type != types.Ask {
		textToLog = " Text: " + command.Text
	}

	shared.Log(types.COMMAND, "Type: "+string(command.Type)+textToLog)

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

func (s *Server) initWave() {

}

// TODO : Refactor to reset topology and neighbors then to fit the second algo later
// handleWaveCount
func (s *Server) handleWaveCount(text string) {
	<-textProcessedChan

	s.initWave()

	s.Text = text
	s.countLetterOccurrences(text)

	shared.Log(types.WAVE, shared.ORANGE+"Start building topology..."+shared.RESET)

	// TODO: Still usefull ?
	// Tri des clés de la map pour les envoyer dans l'ordre
	var keysNeighbors []int
	for key := range s.Neighbors {
		keysNeighbors = append(keysNeighbors, key)
	}
	sort.Ints(keysNeighbors)

	iteration := 1

	for len(s.Counts) < s.NbProcesses {

		shared.Log(types.WAVE, shared.PINK+"Iteration "+strconv.Itoa(iteration)+shared.RESET)
		iteration++

		message := types.Message{
			Counts: s.Counts,
			Number: s.Number,
			Active: true,
		}

		for _, key := range keysNeighbors {
			err := s.sendWaveMessage(message, s.Neighbors[key])
			if err != nil {
				shared.Log(types.ERROR, err.Error())
			}
			shared.Log(types.WAVE, "Sent message to P"+strconv.Itoa(key))
		}

		for _, key := range keysNeighbors {
			message := <-waveMessageChans[key]
			shared.Log(types.WAVE, "Received message from P"+strconv.Itoa(key))
			for letter, count := range message.Counts {
				s.Counts[letter] = count
			}
			if !message.Active {
				delete(s.ActiveNeighbors, message.Number)
			}
		}
	}
	shared.Log(types.WAVE, shared.ORANGE+"Topology built!"+shared.RESET)

	message := types.Message{
		Counts: s.Counts,
		Number: s.Number,
		Active: false,
	}

	var keysActiveNeighbors []int
	for key := range s.ActiveNeighbors {
		keysActiveNeighbors = append(keysActiveNeighbors, key)
	}
	sort.Ints(keysActiveNeighbors)

	// TODO: Mettre l'envoi et la purge pour un processus donné à la suite ?
	shared.Log(types.WAVE, "Sending final counts map to remaining active neighbors...")
	// Envoi de la map de compteurs aux voisins actifs
	for _, key := range keysActiveNeighbors {
		err := s.sendWaveMessage(message, s.ActiveNeighbors[key])
		if err != nil {
			shared.Log(types.ERROR, err.Error())
		}
		shared.Log(types.WAVE, "Sent message to P"+strconv.Itoa(key))
	}

	shared.Log(types.WAVE, "Purging messages from remaining active neighbors...")
	// Purge des derniers messages reçus
	for _, key := range keysActiveNeighbors {
		<-waveMessageChans[key]
		shared.Log(types.WAVE, "Purged message from P"+strconv.Itoa(key))
	}
	shared.Log(types.WAVE, shared.CYAN+"Counts: "+fmt.Sprint(s.Counts)+shared.RESET)
	shared.Log(types.INFO, "Text "+text+" has been processed.")
	textProcessedChan <- true
}

// handleProbeCount
func (s *Server) handleProbeCount(text string) string {
	// TODO
	return ""
}

func (s *Server) handleAsk(text string) string {
	if !<-textProcessedChan {
		return "No text processed yet."
	}

	textProcessedChan <- true
	return s.displayOccurrences(s.Counts)
}

func (s *Server) displayOccurrences(counts map[string]int) string {
	var result string
	result += "---------------------\n"

	result += "Servers in this network can process the following letters: "

	for _, server := range s.Servers {
		result += server.Letter + " "
	}

	result += "\nOccurrences of processable letters in " + s.Text + ":\n"

	for letter, count := range counts {
		if count != 0 {
			result += letter + " : " + strconv.Itoa(count) + "\n"
		}
	}
	result += "---------------------"
	return result
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
