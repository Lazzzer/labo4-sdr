package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/Lazzzer/labo4-sdr/internal/shared"
	"github.com/Lazzzer/labo4-sdr/internal/shared/types"
)

var waveMessageChans = make(map[int](chan types.WaveMessage)) // Map de channels qui gère les messages wave pour chaque processus
var probeEchoMessageChans = make(map[int](chan types.ProbeEchoMessage))
var textProcessedChan = make(chan bool, 1) // Channel qui gère la fin du traitement du texte
var emitterChan = make(chan bool, 1)       // Channel qui gère le fait que le serveur est émetteur ou non

// TODO : Refactor to have subservers for each algo
type Server struct {
	// Network
	Address string               `json:"address"` // Adresse du serveur
	Servers map[int]types.Server `json:"servers"` // Map des serveurs disponibles dans le réseau
	// Process
	Number          int                  `json:"number"`           // Numéro du processus
	NbProcesses     int                  `json:"nb_processes"`     // Nombre total de processus
	Letter          string               `json:"letter"`           // Lettre gérée par le processus pour le comptage des occurrences
	Parent          int                  `json:"parent"`           // Numéro du processus parent
	Neighbors       map[int]types.Server `json:"neighbors"`        // Map prenant en clé le numéro du processus voisin et en valeur ses infos pour la communication
	ActiveNeighbors map[int]bool         `json:"active_neighbors"` // Map prenant en clé le numéro du processus voisin et en valeur un booléen
	Counts          map[string]int       `json:"counts"`           // Map prenant en clé la lettre gérée par un processus et en valeur le nombre d'occurrences
	Text            string               `json:"text"`             // Texte reçu par le serveur
}

// Init est la fonction principale d'initialisation du serveur qui se lance au démarrage du programme.
// Elle utilise une liste d'adjacence valide pour créer un graphe logique des serveurs présents dans le réseau.
func (s *Server) Init(adjacencyList *map[int][]int) {
	textProcessedChan <- false
	emitterChan <- false

	// Initialisation de la map des voisins avec la liste d'adjacence
	s.Neighbors = make(map[int]types.Server)
	for i := 0; i < len((*adjacencyList)[s.Number]); i++ {
		s.Neighbors[(*adjacencyList)[s.Number][i]] = s.Servers[(*adjacencyList)[s.Number][i]]
		waveMessageChans[(*adjacencyList)[s.Number][i]] = make(chan types.WaveMessage, 1)
		probeEchoMessageChans[(*adjacencyList)[s.Number][i]] = make(chan types.ProbeEchoMessage, 1)
	}
}

func (s *Server) Run() {
	connection := s.startListening()
	defer connection.Close()

	shared.Log(types.INFO, shared.GREEN+"Server #"+strconv.Itoa(s.Number)+" as Process P"+strconv.Itoa(s.Number)+" listening on "+s.Address+shared.RESET)

	s.handleCommunications(connection)
}

// init permet l'initialisation des variables du serveur en fonction du type d'algorithme utilisé et (ré)initialise la
// map de compteurs, le texte et la map des voisins actifs pour l'algorithme ondulatoire.
func (s *Server) init(isWave bool) {
	s.Text = ""
	s.Counts = make(map[string]int)

	if isWave {
		s.ActiveNeighbors = make(map[int]bool)
		for i := range s.Neighbors {
			s.ActiveNeighbors[i] = true
		}
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
	for {
		buffer := make([]byte, 1024)
		n, addr, err := connection.ReadFromUDP(buffer)
		if err != nil {
			shared.Log(types.ERROR, err.Error())
			continue
		}
		communication := string(buffer[0:n])
		err = s.handleProbeEchoMessage(communication)
		if err == nil {
			continue
		}
		err = s.handleWaveMessage(communication)
		if err == nil {
			continue
		}
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

func (s *Server) countLetterOccurrences(text string) {
	s.Counts[s.Letter] = strings.Count(strings.ToUpper(text), s.Letter)
	shared.Log(types.INFO, "Letter "+s.Letter+" found "+strconv.Itoa(s.Counts[s.Letter])+" time(s) in "+text)
}

// handleWaveMessage gère les messages reçus des autres serveurs en UDP.
func (s *Server) handleWaveMessage(messageStr string) error {
	message, err := shared.Parse[types.WaveMessage](messageStr)

	if err != nil || message.Number == 0 {
		return fmt.Errorf("invalid message type")
	}

	go func() {
		waveMessageChans[message.Number] <- *message
	}()

	return nil
}

func (s *Server) handleProbeEchoMessage(messageStr string) error {
	message, err := shared.Parse[types.ProbeEchoMessage](messageStr)
	if err == nil {
		if message.Type == types.Probe || message.Type == types.Echo {
			go func() {
				probeEchoMessageChans[message.Number] <- *message
			}()

			if <-emitterChan {
				emitterChan <- true
				return nil
			}

			go func() {
				<-textProcessedChan

				emitterChan <- true

				s.init(false)

				receivedMessage := <-probeEchoMessageChans[message.Number]
				shared.Log(types.PROBE, "message received from P"+strconv.Itoa(receivedMessage.Number))

				s.Text = *receivedMessage.Text
				s.countLetterOccurrences(s.Text)
				s.Parent = receivedMessage.Number

				message := types.ProbeEchoMessage{
					Type:   types.Probe,
					Number: s.Number,
					Text:   &s.Text,
				}

				shared.Log(types.PROBE, "Sending probe to neighbors...")
				for i := range s.Neighbors {
					if i != s.Parent {
						s.sendProbeEchoMessage(message, s.Neighbors[i])
						shared.Log(types.PROBE, "Probe sent to P"+strconv.Itoa(i))
					}
				}
				shared.Log(types.PROBE, "Waiting messages from neighbors...")
				for i := range s.Neighbors {
					if i == s.Parent {
						continue
					}
					message := <-probeEchoMessageChans[i]
					shared.Log(types.PROBE, "Received message from P"+strconv.Itoa(i))
					if message.Type == types.Echo {
						shared.Log(types.ECHO, "Received echo from P"+strconv.Itoa(i))
						for letter, count := range *message.Counts {
							s.Counts[letter] = count
						}
					}
				}

				message = types.ProbeEchoMessage{
					Type:   types.Echo,
					Number: s.Number,
					Counts: &s.Counts,
				}
				s.sendProbeEchoMessage(message, s.Neighbors[s.Parent])
				shared.Log(types.ECHO, "Echo sent to P"+strconv.Itoa(s.Parent))

				shared.Log(types.PROBE, shared.CYAN+"Counts: "+fmt.Sprint(s.Counts)+shared.RESET)
				shared.Log(types.INFO, "Text "+s.Text+" has been processed.")
				textProcessedChan <- true
				<-emitterChan
				emitterChan <- false
			}()
			return nil
		}
	}
	return fmt.Errorf("invalid message type")
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

	if command.Type == types.Ask {
		return s.handleAsk(command.Text), nil
	}

	<-textProcessedChan
	switch command.Type {
	case types.WaveCount:
		s.init(true)
		s.handleWaveCount(command.Text)
	case types.ProbeCount:
		s.init(false)
		return s.handleProbeCount(command.Text), nil
	}
	return "", nil
}

// handleWaveCount
func (s *Server) handleWaveCount(text string) {
	s.Text = text
	s.countLetterOccurrences(text)

	shared.Log(types.WAVE, shared.ORANGE+"Start building topology..."+shared.RESET)

	iteration := 1
	for len(s.Counts) < s.NbProcesses {
		shared.Log(types.WAVE, shared.PINK+"Iteration "+strconv.Itoa(iteration)+shared.RESET)
		iteration++

		message := types.WaveMessage{
			Counts: s.Counts,
			Number: s.Number,
			Active: true,
		}

		for i, neighbor := range s.Neighbors {
			err := s.sendWaveMessage(message, neighbor)
			if err != nil {
				shared.Log(types.ERROR, err.Error())
			}
			shared.Log(types.WAVE, "Sent message to P"+strconv.Itoa(i))
		}

		for i := range s.Neighbors {
			message := <-waveMessageChans[i]
			shared.Log(types.WAVE, "Received message from P"+strconv.Itoa(i))
			for letter, count := range message.Counts {
				s.Counts[letter] = count
			}
			if !message.Active {
				delete(s.ActiveNeighbors, message.Number)
			}
		}
	}
	shared.Log(types.WAVE, shared.ORANGE+"Topology built!"+shared.RESET)

	message := types.WaveMessage{
		Counts: s.Counts,
		Number: s.Number,
		Active: false,
	}

	shared.Log(types.WAVE, "Sending final counts map to remaining active neighbors...")
	// Envoi de la map de compteurs aux voisins actifs
	for i := range s.ActiveNeighbors {
		err := s.sendWaveMessage(message, s.Neighbors[i])
		if err != nil {
			shared.Log(types.ERROR, err.Error())
		}
		shared.Log(types.WAVE, "Sent message to P"+strconv.Itoa(i))
	}

	shared.Log(types.WAVE, "Purging messages from remaining active neighbors...")
	// Purge des derniers messages reçus
	for i := range s.ActiveNeighbors {
		<-waveMessageChans[i]
		shared.Log(types.WAVE, "Purged message from P"+strconv.Itoa(i))
	}
	shared.Log(types.WAVE, shared.CYAN+"Counts: "+fmt.Sprint(s.Counts)+shared.RESET)
	shared.Log(types.INFO, "Text "+text+" has been processed.")
	textProcessedChan <- true
}

// handleProbeCount
func (s *Server) handleProbeCount(text string) string {
	<-emitterChan
	emitterChan <- true

	s.Parent = s.Number
	s.Text = text
	s.countLetterOccurrences(text)

	message := types.ProbeEchoMessage{
		Type:   types.Probe,
		Number: s.Number,
		Text:   &text,
		Counts: nil,
	}

	shared.Log(types.PROBE, "Sending probe to neighbors...")
	for i, neighbor := range s.Neighbors {
		err := s.sendProbeEchoMessage(message, neighbor)
		if err != nil {
			shared.Log(types.ERROR, err.Error())
		}
		shared.Log(types.PROBE, "Sent probe to P"+strconv.Itoa(i))
	}

	shared.Log(types.PROBE, "Waiting echoes from neighbors...")
	for i := range s.Neighbors {
		message := <-probeEchoMessageChans[i]
		if message.Type == types.Echo {
			shared.Log(types.ECHO, "Received echo from P"+strconv.Itoa(message.Number))
			for letter, count := range *message.Counts {
				s.Counts[letter] = count
			}
		}
	}

	shared.Log(types.WAVE, shared.CYAN+"Counts: "+fmt.Sprint(s.Counts)+shared.RESET)
	shared.Log(types.INFO, "Text "+text+" has been processed.")
	textProcessedChan <- true
	<-emitterChan
	emitterChan <- false

	return s.displayOccurrences(s.Counts)

}

func (s *Server) handleAsk(text string) string {
	if !<-textProcessedChan {
		textProcessedChan <- false
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

func (s *Server) sendWaveMessage(message types.WaveMessage, neighbor types.Server) error {
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

func (s *Server) sendProbeEchoMessage(message types.ProbeEchoMessage, neighbor types.Server) error {
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
