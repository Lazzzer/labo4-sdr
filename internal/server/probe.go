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

// initProbeEchoCountAsRoot initialise le traitement d'un texte avec l'algorithme sondes et échos en tant que processus racine.
// La méthode retourne le résultat pour être traité comme réponse à la commande du client.
func (s *Server) initProbeEchoCountAsRoot(text string) string {
	<-emitterChan
	emitterChan <- true // ainsi, dans le handle, le serveur saura qu'il a déjà émis et qu'il ne doit pas initier l'algorithme de nouveau

	shared.Log(types.PROBE, "Processing text \""+text+"\" as root process")

	s.init(false)
	s.Parent = s.Number
	s.Text = text
	s.countLetterOccurrences(text)

	// Envoi des sondes aux voisins

	message := types.ProbeEchoMessage{
		Type:   types.Probe,
		Number: s.Number,
		Text:   &text,
		Counts: nil,
	}

	for i, neighbor := range s.Neighbors {
		err := sendMessage(message, neighbor)
		if err != nil {
			shared.Log(types.ERROR, err.Error())
		}
		shared.Log(types.PROBE, "Sent probe to P"+strconv.Itoa(i))
	}

	// Attente des réponses des voisins et traitement des échos

	shared.Log(types.ECHO, "Waiting echoes from children...")
	for i := range s.Neighbors {
		message := <-probeEchoMessageChans[i]
		if message.Type == types.Echo {
			shared.Log(types.ECHO, "Received echo from P"+strconv.Itoa(message.Number))
			for letter, count := range *message.Counts {
				s.Counts[letter] = count
			}
		}
	}

	shared.Log(types.INFO, shared.CYAN+"Counts: "+fmt.Sprint(s.Counts)+shared.RESET)
	shared.Log(types.INFO, "Text \""+text+"\" has been processed")
	textProcessedChan <- true
	<-emitterChan
	emitterChan <- false

	return s.displayOccurrences(s.Counts)
}

// initProbeEchoCountAsLeaf initialise le traitement d'un texte avec l'algorithme sondes et échos en tant que processus feuille.
func (s *Server) initProbeEchoCountAsLeaf(message types.ProbeEchoMessage) {
	<-textProcessedChan

	emitterChan <- true

	s.init(false)

	receivedMessage := <-probeEchoMessageChans[message.Number]
	shared.Log(types.PROBE, "Received Probe from P"+strconv.Itoa(receivedMessage.Number))
	shared.Log(types.PROBE, "Processing text \""+*receivedMessage.Text+"\" as leaf process")

	s.Text = *receivedMessage.Text
	s.countLetterOccurrences(s.Text)
	s.Parent = receivedMessage.Number

	// Envoi d'une sonde à tous les voisins sauf au parent

	newMessage := types.ProbeEchoMessage{
		Type:   types.Probe,
		Number: s.Number,
		Text:   &s.Text,
	}

	for i := range s.Neighbors {
		if i != s.Parent {
			sendMessage(newMessage, s.Neighbors[i])
			shared.Log(types.PROBE, "Sent probe to P"+strconv.Itoa(i))
		}
	}

	// Attente des réponses des voisins et traitement des échos

	for i := range s.Neighbors {
		if i == s.Parent {
			continue
		}
		message := <-probeEchoMessageChans[i]
		if message.Type == types.Echo {
			shared.Log(types.ECHO, "Received echo from P"+strconv.Itoa(i))
			for letter, count := range *message.Counts {
				s.Counts[letter] = count
			}
		} else {
			shared.Log(types.PROBE, "Received probe from P"+strconv.Itoa(i)+", not handling it")
		}
	}

	// Envoi de l'écho au parent

	newMessage = types.ProbeEchoMessage{
		Type:   types.Echo,
		Number: s.Number,
		Counts: &s.Counts,
	}
	sendMessage(newMessage, s.Neighbors[s.Parent])
	shared.Log(types.ECHO, "Sent echo to P"+strconv.Itoa(s.Parent))

	shared.Log(types.INFO, shared.CYAN+"Counts: "+fmt.Sprint(s.Counts)+shared.RESET)
	shared.Log(types.INFO, "Processed text \""+s.Text+"\" as leaf process, root process can now display the result")
	textProcessedChan <- false // Les serveurs feuilles ne peuvent pas répondre à des asks car leur map de comptage n'est pas complète
	<-emitterChan
	emitterChan <- false
}

// handleProbeEchoMessage traite un message de type Probe ou Echo.
// Si le serveur n'a pas encore émis, il initie l'algorithme en tant que processus feuille dans une goroutine.
func (s *Server) handleProbeEchoMessage(messageStr string) error {
	message, err := shared.Parse[types.ProbeEchoMessage](messageStr)
	if err == nil {
		if message.Type == types.Probe || message.Type == types.Echo {
			go func() {
				probeEchoMessageChans[message.Number] <- *message
			}()
			if <-emitterChan {
				emitterChan <- true // si le serveur a déjà émis, il ne doit pas initier l'algorithme de nouveau
				return nil
			}
			go s.initProbeEchoCountAsLeaf(*message)
			return nil
		}
	}
	return fmt.Errorf("invalid message type")
}
