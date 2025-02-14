package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	// Assurez-vous d'importer correctement vos modèles
)

func (aw AppWrapper) GetAllPostByCat(w http.ResponseWriter, r *http.Request) {
	// Vérifiez que la méthode HTTP est GET
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Extraire le nom de la catégorie de l'URL
	prefix := "/category/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		http.NotFound(w, r)
		return
	}
	nameCat := strings.TrimPrefix(r.URL.Path, prefix)
	nameCat = strings.TrimSpace(nameCat) // Supprimer les espaces inutiles

	if nameCat == "" {
		http.Error(w, "Nom de la catégorie manquant", http.StatusBadRequest)
		return
	}

	userCookie, err := r.Cookie("userID")
	var sessionID string
	if err != nil {
		// If the cookie doesn't exist, sessionID remains an empty string
		sessionID = ""
	} else {
		// The cookie exists; safely access userCookie.Value
		sessionID = userCookie.Value
	}

	var username string

	if sessionID != "" {
		// Retrieve the username from the session ID
		username, err = aw.App.Sessions.GetUsername2(sessionID)
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		username = ""
	}

	// Récupérer les posts par nom de catégorie
	posts, err := aw.App.Category.GetPostsByCategoryName(nameCat, sessionID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des posts", http.StatusInternalServerError)
		return
	}

	// Récupérer toutes les catégories pour l'affichage (par exemple, pour un menu de navigation)
	categories, err := aw.App.Category.GetAllCategory()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des catégories", http.StatusInternalServerError)
		return
	}

	currentCategory, err := aw.App.Category.GetCategoryByName(nameCat)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération de la catégorie", http.StatusInternalServerError)
		return
	}

	// Préparer les données à passer au template
	data := map[string]interface{}{
		"category":   currentCategory, // Catégorie courante
		"posts":      posts,           // Liste des posts
		"categories": categories,      // Toutes les catégories
		"username":   username,        // Nom d'utilisateur connecté
	}

	templatePath := filepath.Join(projectPath, "templates", "page.categoryname.html")
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Execute the template with the data
	err = t.Execute(w, data)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
	}
}
