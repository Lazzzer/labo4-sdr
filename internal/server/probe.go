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
		err := sendMessage(message, neighbor)
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
						sendMessage(message, s.Neighbors[i])
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
				sendMessage(message, s.Neighbors[s.Parent])
				shared.Log(types.ECHO, "Echo sent to P"+strconv.Itoa(s.Parent))

				shared.Log(types.PROBE, shared.CYAN+"Counts: "+fmt.Sprint(s.Counts)+shared.RESET)
				shared.Log(types.INFO, "Text "+s.Text+" has been processed.")
				textProcessedChan <- false
				<-emitterChan
				emitterChan <- false
			}()
			return nil
		}
	}
	return fmt.Errorf("invalid message type")
}
