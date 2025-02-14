package handlers

import (
	"errors"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Profile gère l'affichage du profil utilisateur.
func (aw AppWrapper) Profile(w http.ResponseWriter, r *http.Request) {
	// Extraire le nom d'utilisateur de l'URL
	username := r.URL.Path[len("/profile/"):]

	// Récupérer les informations de l'utilisateur du profil
	userID, user, role, picture, err := aw.App.User.GetByUsername(username)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	sessionUserID, err := r.Cookie("userID")
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusUnauthorized, "Missing user cookie")
		return
	}
	sessionUseriD := sessionUserID.Value

	// Récupérer le nom d'utilisateur et l'ID de l'utilisateur connecté
	var currentUsername string
	var currentUserID string
	sessionCookie, err := r.Cookie("userID")
	if err == nil {
		sessionID := sessionCookie.Value
		currentUsername, _ = aw.App.Sessions.GetUsername2(sessionID)
		currentUserID, _ = aw.App.Sessions.GetUserID(sessionID)
	}

	// Récupérer tous les posts de l'utilisateur avec les informations de likes/dislikes
	posts, err := aw.App.Posts.AllPostByUserProfile(userID, currentUserID, sessionUseriD)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Préparer les données pour le template
	data := map[string]interface{}{
		"User": map[string]interface{}{
			"ID":       userID,
			"Username": user,
			"Picture":  picture,
			"Roles":    role,
		},
		"Posts":           posts,
		"CurrentUsername": currentUsername,
		"LoggedIn":        currentUsername != "",
	}

	// Définir le chemin du template
	templatePath := filepath.Join("templates", "page.profile.html")
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Exécuter le template avec les données
	err = t.Execute(w, data)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
	}
}

func (aw AppWrapper) EditProfile(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Path[len("/profile/edit/"):]

	sessionUser, err := r.Cookie("userID")
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusUnauthorized, "Missing user cookie")
		return
	}

	userID := sessionUser.Value
	profileUserID, err := aw.App.User.GetUserIdByUsername(username)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, "Error retrieving user ID")
		return
	}

	if userID != profileUserID {
		aw.ErrorHandler(w, r, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	if r.Method == http.MethodGet {
		user, email, picture, err := aw.App.User.GetAllInfoUser(profileUserID)
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, "Error retrieving user")
			return
		}

		templatePath := filepath.Join(projectPath, "templates", "page.editprofile.html")

		t, err := template.ParseFiles(templatePath)
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}

		err = t.Execute(w, map[string]interface{}{"user": user, "email": email, "picture": picture})
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	if r.Method == http.MethodPost {
		err := r.ParseMultipartForm(20 << 20)
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusBadRequest, "Invalid form data")
			return
		}

		var avatarPath string
		file, handler, err := r.FormFile("image")
		if err != nil && !errors.Is(err, http.ErrMissingFile) {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, "Error processing file")
			return
		}

		if file != nil {
			defer file.Close()

			validExtensions := map[string]bool{".jpeg": true, ".jpg": true, ".png": true, ".gif": true}
			ext := filepath.Ext(handler.Filename)
			if !validExtensions[ext] {
				aw.ErrorHandler(w, r, http.StatusBadRequest, "Invalid image type")
				return
			}

			avatarPath = filepath.Join(projectPath, "static", "images_profile", handler.Filename)

			dst, err := os.Create(avatarPath)
			if err != nil {
				aw.ErrorHandler(w, r, http.StatusInternalServerError, "Error saving avatar")
				return
			}
			defer dst.Close()

			if _, err := io.Copy(dst, file); err != nil {
				aw.ErrorHandler(w, r, http.StatusInternalServerError, "Error copying file")
				return
			}

		}

		imageName := "default.jpg"
		if avatarPath != "" {
			imageName = handler.Filename
		}

		// Mise à jour du profil utilisateur
		usernamePro := r.FormValue("username")
		if avatarPath != "" {
			usernamePro, err = aw.App.Sessions.GetUsername2(userID)
			if err != nil {
				aw.ErrorHandler(w, r, http.StatusInternalServerError, "Error retrieving user ID")
				return
			}
		}

		err = aw.App.User.EditProfile(usernamePro, imageName, userID)
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, "Error updating profile")
			return
		}

		newUsername, err := aw.App.Sessions.GetUsername2(userID)
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, "Error retrieving user ID")
			return
		}

		http.Redirect(w, r, "/profile/"+newUsername, http.StatusSeeOther)
		return
	}

	// Méthode non autorisée
	aw.ErrorHandler(w, r, http.StatusMethodNotAllowed, "Method not allowed")
}
