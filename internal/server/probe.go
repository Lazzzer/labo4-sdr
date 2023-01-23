package server

import (
	"encoding/json"
	"fmt"
	"net"
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
				textProcessedChan <- false
				<-emitterChan
				emitterChan <- false
			}()
			return nil
		}
	}
	return fmt.Errorf("invalid message type")
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
