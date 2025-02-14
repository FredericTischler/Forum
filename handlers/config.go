// config.go
package handlers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	GithubClientID     string `json:"github_client_id"`
	GithubClientSecret string `json:"github_client_secret"`
	GoogleClientID     string `json:"google_client_id"`
	GoogleClientSecret string `json:"google_client_secret"`
}

var AppConfig Config

func LoadConfig() {
	log.Println("Début du chargement de la configuration...")
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Printf("Erreur lors de la lecture du fichier config.json: %v", err)
		log.Fatal("Chemin actuel:", getCurrentPath())
		return
	}

	log.Printf("Contenu du fichier config.json: %s", string(file))

	err = json.Unmarshal(file, &AppConfig)
	if err != nil {
		log.Fatal("Erreur lors du parsing du fichier config.json:", err)
		return
	}

	log.Printf("Configuration chargée: %+v", AppConfig)
}

func getCurrentPath() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("Erreur lors de la récupération du chemin: %v", err)
		return ""
	}
	return dir
}
