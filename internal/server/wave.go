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
	"fmt"
	"strconv"

	"github.com/Lazzzer/labo4-sdr/internal/shared"
	"github.com/Lazzzer/labo4-sdr/internal/shared/types"
)

// initiateWaveCount initialise le comptage des occurrences de la lettre du serveur et applique l'algorithme ondulatoire
// pour transmettre les informations aux voisins et recevoir leur comptage.
func (s *Server) initiateWaveCount(text string) {
	s.init(true)
	s.Text = text
	s.countLetterOccurrences(text)

	shared.Log(types.WAVE, shared.ORANGE+"Start building topology..."+shared.RESET)

	// Boucle de création de la topologie

	iteration := 1
	for len(s.Counts) < s.NbProcesses {
		shared.Log(types.WAVE, shared.PINK+"Iteration "+strconv.Itoa(iteration)+shared.RESET)
		iteration++

		message := types.WaveMessage{
			Type:   types.Wave,
			Counts: s.Counts,
			Number: s.Number,
			Active: true,
		}
		for i, neighbor := range s.Neighbors {
			err := sendMessage(message, neighbor)
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
				shared.Log(types.WAVE, "P"+strconv.Itoa(i)+" is now inactive.")
			}
		}
	}
	shared.Log(types.WAVE, shared.ORANGE+"Topology built!"+shared.RESET)

	// Envoi du message final aux voisins actifs

	message := types.WaveMessage{
		Type:   types.Wave,
		Counts: s.Counts,
		Number: s.Number,
		Active: false,
	}

	for i := range s.ActiveNeighbors {
		err := sendMessage(message, s.Neighbors[i])
		if err != nil {
			shared.Log(types.ERROR, err.Error())
		}
		shared.Log(types.WAVE, "Sent final message to active process P"+strconv.Itoa(i))
	}

	// Purge des derniers messages reçus

	for i := range s.ActiveNeighbors {
		<-waveMessageChans[i]
		shared.Log(types.WAVE, "Purged message from P"+strconv.Itoa(i))
	}

	shared.Log(types.WAVE, shared.CYAN+"Counts: "+fmt.Sprint(s.Counts)+shared.RESET)
	shared.Log(types.INFO, "Text "+text+" has been processed.")
	textProcessedChan <- true
}

// handleWaveMessage gère les messages reçus des autres serveurs en UDP et s'assure que le message est destiné à l'algorithme ondulatoire
// en vérifiant le type du message.
func (s *Server) handleWaveMessage(messageStr string) error {
	message, err := shared.Parse[types.WaveMessage](messageStr)

	if err != nil || message.Type != types.Wave {
		return fmt.Errorf("invalid message type")
	}

	go func() {
		waveMessageChans[message.Number] <- *message
	}()

	return nil
}
