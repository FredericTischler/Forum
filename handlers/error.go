package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)

func (aw AppWrapper) ErrorHandler(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	templateError := filepath.Join(projectPath, "templates", "error.page.html")
	tmpl, err := template.ParseFiles(templateError)
	if err != nil {
		log.Println("Erreur lors du chargement du modèle :", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Title   string
		Code    int
		Message string
	}{
		Title:   "Erreur " + strconv.Itoa(statusCode),
		Code:    statusCode,
		Message: message,
	}

	w.WriteHeader(statusCode)
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Erreur lors de l'exécution du modèle :", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
