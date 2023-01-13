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
)

type CommandType string // Type de commande

const (
	Count CommandType = "count" // Commande de comptage des occurrences de lettres
	Ask   CommandType = "ask"   // Commande de demande du résultat d'un comptage sur un mot
	Quit  CommandType = "quit"  // Commande de fermeture du client

)

// Command représente une commande envoyée par un client.
type Command struct {
	Type   CommandType `json:"command_type"`    // Type de la commande
	Text   string      `json:"text,omitempty"`  // Texte à analyser
	Server *int        `json:"value,omitempty"` // Nb de serveurs à utiliser
}
