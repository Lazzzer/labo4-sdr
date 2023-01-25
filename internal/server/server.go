// Auteurs: Jonathan Friedli, Lazar Pavicevic
// Labo 4 SDR

// Package serveur propose un serveur UDP connecté dans un réseau de serveurs. Le serveur peut recevoir des commandes de clients UDP et
// traiter des occurrences de lettre dans des textes de manière distribuée en utilisant l'algorithme ondulatoire ou l'algorithme sondes et échos.
// Il est possible de choisir l'algorithme à utiliser en lui envoyant la commande correspondante avec le texte à traiter.
// Le résultat est communiqué soit directement au client émetteur de la commande avec l'algorithme sondes et échos, soit sur demande
// avec une commande "ask" lors de l'utilisation de l'algorithme ondulatoire. De plus, dans une analyse utilisant l'algorithme sondes et échos, le processus racine peut également recevoir
// des commandes "ask" tant qu'il n'y a pas eu de nouveau traitement de texte.
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

// Channels
var waveMessageChans = make(map[int](chan types.WaveMessage))           // Map de channels qui gère les messages issus de l'algorithme ondulatoire pour chaque processus voisin
var probeEchoMessageChans = make(map[int](chan types.ProbeEchoMessage)) // Map de channels qui gère les messages issus de l'algorithme sondes et échos pour chaque processus voisin
var textProcessedChan = make(chan bool, 1)                              // Channel qui indique si un texte a été traité ou non et bloque le traitement simultané
var emitterChan = make(chan bool, 1)                                    // Channel qui gère si le serveur a déjà émis un message dans l'algorithme sondes et échos

// Server est la structure qui représente un serveur UDP connecté dans un réseau de serveurs.
// Elle contient les propriétés du processus et les propriétés du réseau.
type Server struct {
	// Propriétés liées au réseau

	Address string               `json:"address"` // Adresse du serveur
	Servers map[int]types.Server `json:"servers"` // Map des serveurs disponibles dans le réseau, la clé est le numéro de processus

	// Propriétés du processus

	Number          int                  `json:"number"`           // Numéro du processus
	NbProcesses     int                  `json:"nb_processes"`     // Nombre total de processus
	Letter          string               `json:"letter"`           // Lettre gérée par le processus pour le comptage des occurrences
	Parent          int                  `json:"parent"`           // Numéro du processus parent pour l'algorithme sondes et échos
	Neighbors       map[int]types.Server `json:"neighbors"`        // Map prenant en clé le numéro du processus voisin et en valeur ses infos pour la communication sur le ré
	ActiveNeighbors map[int]bool         `json:"active_neighbors"` // Map prenant en clé le numéro du processus voisin et en valeur un booléen pour l'algorithme ondulatoire
	Counts          map[string]int       `json:"counts"`           // Map prenant en clé la lettre gérée par un processus et en valeur le nombre d'occurrences
	Text            string               `json:"text"`             // Texte à traiter reçu par le serveur
}

// Init est la fonction principale d'initialisation du serveur qui se lance au démarrage du programme.
// Elle utilise une liste d'adjacence valide représentant un graphe logique des serveurs présents dans le réseau.
// La méthode initialise les maps de voisin du processus ainsi que les channels de communication avec les voisins.
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

// Run permet de démarrer l'écoute des connexions entrantes sur le port du serveur.
// et lance la méthode principale qui boucle sur les connexions entrantes.
func (s *Server) Run() {
	udpAddr, err := net.ResolveUDPAddr("udp4", s.Address)
	if err != nil {
		log.Fatal(err)
	}

	connection, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()

	shared.Log(types.INFO, shared.GREEN+"Process P"+strconv.Itoa(s.Number)+" listening on "+s.Address+shared.RESET)

	s.handleCommunications(connection)
}

// init permet l'initialisation des variables du serveur en fonction du type d'algorithme utilisé et (ré)initialise la
// map de compteurs et la map des voisins actifs pour l'algorithme ondulatoire.
func (s *Server) init(isWave bool) {
	s.Counts = make(map[string]int)

	if isWave {
		s.ActiveNeighbors = make(map[int]bool)
		for i := range s.Neighbors {
			s.ActiveNeighbors[i] = true
		}
	}
}

// handleCommunications gère les communications du serveur.
// La méthode écoute les messages entre serveurs pour les deux algorithmes  ainsi que les commandes des clients.
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

		// S'il ne s'agit pas d'un message pour l'exécution d'un algorithme, on traite une commande dans un goroutine
		go func() {
			response, err := s.handleCommand(communication)
			if err != nil {
				shared.Log(types.ERROR, err.Error())
			}
			// Envoi de la réponse à l'adresse du client seulement si le serveur a généré un message de réponse
			if response != "" {
				_, err = connection.WriteToUDP([]byte(response), addr)
				if err != nil {
					shared.Log(types.ERROR, err.Error())
				}
				shared.Log(types.INFO, "Response sent to "+addr.String())
			}
		}()
	}
}

// handleCommand gère les commandes reçues des clients UDP.
// Si la commande est valide, on traite le type de commande et on retourne un message de réponse si nécessaire.
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
		s.initWaveCount(command.Text)
	case types.ProbeCount:
		return s.initProbeEchoCountAsRoot(command.Text), nil
	}
	return "", nil
}

// handleAsk gère la commande "ask" des clients UDP. Si le serveur a déjà traité un texte, on retourne le nombre d'occurrences
// de la lettre du serveur dans le texte. Sinon, on retourne un message d'erreur.
func (s *Server) handleAsk(text string) string {
	if !<-textProcessedChan {
		textProcessedChan <- false
		return "No processed text to show"
	}

	textProcessedChan <- true
	return s.displayOccurrences(s.Counts)
}

// countLetterOccurrences compte le nombre d'occurrences de la lettre du serveur dans le texte passé en paramètre.
func (s *Server) countLetterOccurrences(text string) {
	s.Counts[s.Letter] = strings.Count(strings.ToUpper(text), s.Letter)
	shared.Log(types.INFO, "Letter "+s.Letter+" found "+strconv.Itoa(s.Counts[s.Letter])+" time(s) in "+text)
}

// displayOccurrences retourne une chaîne de caractères contenant le nombre d'occurrences de chaque lettre du texte
// traité par les serveurs du réseau. Seul les lettres capables d'être traitées par un serveur sont affichées.
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

// sendMessage est une fonction générique permettant d'envoyer un message de type T à un voisin du réseau en UDP.
// Le message peut être de type WaveMessage ou ProbeEchoMessage.
func sendMessage[T types.WaveMessage | types.ProbeEchoMessage](message T, neighbor types.Server) error {
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
