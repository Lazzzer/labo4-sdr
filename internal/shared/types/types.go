// Auteurs: Jonathan Friedli, Lazar Pavicevic
// Labo 4 SDR

// Package types propose différents types utilisés par l'application pour parser le fichier de configuration, les messages et les commandes.
package types

// Config représente la configuration du réseau de serveurs.
type Config struct {
	Servers map[int]string `json:"servers"` // Liste des adresse des serveurs disponibles
}

type ServerConfig struct {
	Servers       map[int]Server `json:"servers"`        // Liste des serveurs disponibles avec leur lettre et leur adresse
	AdjacencyList map[int][]int  `json:"adjacency_list"` // Liste d'adjacence des serveurs
}

type Server struct {
	Letter  string `json:"letter"`  // Lettre à compter
	Address string `json:"address"` // Adresse du serveur
}

type LogType string // Type de log

const (
	INFO    LogType = "INFO"    // Log d'information
	DEBUG   LogType = "DEBUG"   // Log de debug
	ERROR   LogType = "ERROR"   // Log d'erreur
	COMMAND LogType = "COMMAND" // Log de commande
	WAVE    LogType = "WAVE"    // Log de message wave
	PROBE   LogType = "PROBE"   // Log de message probe
	ECHO    LogType = "ECHO"    // Log de message echo
)

type CommandType string // Type de commande

const (
	WaveCount  CommandType = "wave"  // Commande de comptage des occurrences de lettres avec un algorithme ondulatoire
	ProbeCount CommandType = "probe" // Commande de comptage des occurrences de lettres avec un algorithme de sondes et échos
	Ask        CommandType = "ask"   // Commande de demande du résultat d'un comptage sur un texte
	Quit       CommandType = "quit"  // Commande de fermeture du client
)

// Command représente une commande envoyée par un client.
type Command struct {
	Type CommandType `json:"command_type"`   // Type de la commande
	Text string      `json:"text,omitempty"` // Texte à analyser
}

type MessageType string // Type de message probe ou echo

// WaveMessage représente un message de l'algorithme ondulatoire envoyé par un processus.
type WaveMessage struct {
	Type   MessageType    `json:"type"`   // Type de message
	Counts map[string]int `json:"counts"` //  Map qui contient le compteur de chaque lettre gérée par les processus
	Number int            `json:"number"` // Numéro du processus qui envoie le message
	Active bool           `json:"active"` // Indique si le voisin est actif ou non
}

const (
	Wave  MessageType = "wave"  // Message de type wave
	Probe MessageType = "probe" // Message de type probe
	Echo  MessageType = "echo"  // Message de type echo
)

// ProbeEchoMessage représente un message de l'algorithme de sondes et échos envoyé par un processus.
type ProbeEchoMessage struct {
	Type   MessageType     `json:"type"`   // Type de message (sonde ou écho)
	Number int             `json:"number"` // Numéro du processus qui envoie le message
	Text   *string         `json:"text"`   // Texte à analyser
	Counts *map[string]int `json:"counts"` //  Map qui contient le compteur de chaque lettre gérée par les processus
}
