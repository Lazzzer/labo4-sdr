package server

import "github.com/Lazzzer/labo4-sdr/internal/shared/types"

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

type Process struct {
}

func (s *Server) Init(adjacencyList *map[int][]int) {

	// Initialisation de la map des voisins
	for _, neighbor := range (*adjacencyList)[s.Number] {
		s.Neighbors[neighbor] = s.Servers[neighbor]
	}

	// Initialisation du nombre de voisins
	s.NbNeighbors = len(s.Neighbors)
}

func (s *Server) Run() {

}
