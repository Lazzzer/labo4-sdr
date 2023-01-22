package shared

import (
	"encoding/json"
	"log"

	"github.com/Lazzzer/labo4-sdr/internal/shared/types"
)

// Parse permet de parser un objet JSON en un objet de type T.
func Parse[T types.Config | types.ServerConfig | types.Command | types.WaveMessage | types.ProbeEchoMessage](jsonStr string) (*T, error) {
	var object T

	err := json.Unmarshal([]byte(jsonStr), &object)
	if err != nil {
		return nil, err
	}

	return &object, nil
}

// Log permet d'afficher un message dans la console avec une couleur différente selon le type de log.
func Log(logType types.LogType, message string) {
	switch logType {
	case types.INFO:
		log.Println(CYAN + "(INFO) " + RESET + message)
	case types.DEBUG:
		log.Println(ORANGE + "(DEBUG) " + RESET + message)
	case types.ERROR:
		log.Println(RED + "(ERROR) " + RESET + message)
	case types.COMMAND:
		log.Println(YELLOW + "(COMMAND) " + RESET + message)
	case types.WAVE:
		log.Println(GREEN + "(WAVE) " + RESET + message)
	case types.PROBE:
		log.Println(PURPLE + "(PROBE) " + RESET + message)
	case types.ECHO:
		log.Println(PINK + "(ECHO) " + RESET + message)
	}

}

// Variables pour colorer le texte dans la console
var RESET = "\033[0m"         // Variable pour réinitialiser la couleur du texte
var RED = "\033[31m"          // Variable pour colorer le texte en rouge
var PINK = "\033[38;5;219m"   // Variable pour colorer le texte en rose
var PURPLE = "\033[38;5;198m" // Variable pour colorer le texte en violet

var GREEN = "\033[32m"        // Variable pour colorer le texte en vert
var YELLOW = "\033[33m"       // Variable pour colorer le texte en jaune
var ORANGE = "\033[38;5;208m" // Variable pour colorer le texte en orange
var CYAN = "\033[36m"         // Variable pour colorer le texte en cyan
var BOLD = "\033[1m"          // Variable pour changer le texte en gras
