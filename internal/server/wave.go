package server

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"github.com/Lazzzer/labo4-sdr/internal/shared"
	"github.com/Lazzzer/labo4-sdr/internal/shared/types"
)

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
