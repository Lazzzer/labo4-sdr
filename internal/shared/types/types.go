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
	WaveCount  CommandType = "wave"  // Commande de comptage des occurrences de lettres avec un algorithme ondulatoire
	ProbeCount CommandType = "probe" // Commande de comptage des occurrences de lettres avec un algorithme de sondes et échos
	Ask        CommandType = "ask"   // Commande de demande du résultat d'un comptage sur un mot
	Quit       CommandType = "quit"  // Commande de fermeture du client
)

// Command représente une commande envoyée par un client.
type Command struct {
	Type CommandType `json:"command_type"`   // Type de la commande
	Text string      `json:"text,omitempty"` // Texte à analyser
}

type Message struct {
	Number   int            `json:"number"`   // Numéro du processus qui envoie le message
	Topology map[string]int `json:"topology"` // Map de la topologie qui contient le compteur de chaque lettre
}
