package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
)

func (aw AppWrapper) Notification(w http.ResponseWriter, r *http.Request) {
	// Vérifie si l'utilisateur est authentifié via un cookie
	sessionUser, err := r.Cookie("userID")
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusUnauthorized, "Missing user cookie")
		return
	}

	userID := sessionUser.Value

	// Appelle la méthode pour récupérer les notifications
	notifications, err := aw.App.Notification.GetNotification(userID)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, "Failed to fetch notifications: "+err.Error())
		return
	}

	// Chemin du fichier template
	templatePath := filepath.Join(projectPath, "templates", "page.notification.html")

	// Parse le template
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, "Failed to load template: "+err.Error())
		return
	}

	// Prépare les données pour le template
	data := map[string]interface{}{
		"notifications": notifications,
		"username":      sessionUser.Value, // Exemple : utiliser le cookie pour remplir le nom d'utilisateur si nécessaire
	}

	// Exécute le template avec les données
	err = t.Execute(w, data)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, "Failed to render template: "+err.Error())
	}
}

func (aw AppWrapper) ReadNotification(w http.ResponseWriter, r *http.Request) {
	// Vérifie si l'utilisateur est authentifié via un cookie
	sessionUser, err := r.Cookie("userID")
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusUnauthorized, "Missing user cookie")
		return
	}

	userID := sessionUser.Value

	notificationIDString := r.URL.Path[len("/notification/read/"):]
	notificationID, err := strconv.Atoi(notificationIDString)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	// Appelle la méthode pour marquer la notification comme lue
	err = aw.App.Notification.ReadNotification(userID, notificationID)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, "Failed to mark notification as read: "+err.Error())
		return
	}

	getPostid, err := aw.App.Notification.GetPostId(notificationID)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, "Failed to get post id: "+err.Error())
		return
	}

	getPostidString := strconv.Itoa(getPostid)

	// Redirige l'utilisateur vers la page des notifications
	http.Redirect(w, r, "/post/direct/"+getPostidString, http.StatusSeeOther)

}


