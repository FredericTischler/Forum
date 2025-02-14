package handlers

//Description : Gestion des routes pour l'authentification des utilisateurs (inscription, connexion, OAuth).
//À faire :
//
//    Gérer les requêtes HTTP pour l'inscription et la connexion des utilisateurs.
//    Implémenter les routes pour l'authentification via Google et GitHub OAuth.
//    Gestion des sessions et des cookies d'authentification.
//    Implémenter la déconnexion de l'utilisateur.

import (
	"database/sql"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", "app.db")
	if err != nil {
		log.Fatal(err)
	}

}

// Fonction pour hasher le mot de passe
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// Fonction pour vérifier le mot de passe hashé
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Fonction pour valider le mot de passe
func validatePassword(password string) bool {
	var (
		minLength   = 8
		upperCase   = regexp.MustCompile(`[A-Z]`)
		lowerCase   = regexp.MustCompile(`[a-z]`)
		digit       = regexp.MustCompile(`[0-9]`)
		specialChar = regexp.MustCompile(`[!@#~$%^&*()_+|<>?:{}]`)
	)

	if len(password) < minLength {
		return false
	}
	if !upperCase.MatchString(password) {
		return false
	}
	if !lowerCase.MatchString(password) {
		return false
	}
	if !digit.MatchString(password) {
		return false
	}
	if !specialChar.MatchString(password) {
		return false
	}
	return true
}

// Handler pour afficher la page d'inscription et gérer l'enregistrement
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Chemin du template d'enregistrement
	templateRegister := filepath.Join(projectPath, "templates", "page.register.html")

	if r.Method == "GET" {
		http.ServeFile(w, r, templateRegister)
		return
	}

	if r.Method == "POST" {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Validation des champs
		if username == "" || email == "" || password == "" {
			http.Error(w, "Tous les champs sont obligatoires", http.StatusBadRequest)
			return
		}

		if !validateEmail(email) {
			http.Error(w, "Adresse e-mail invalide, l'email doit être sous cette forme (example@example.validTLD)", http.StatusBadRequest)
			return
		}

		if !validatePassword(password) {
			http.Error(w, "Le mot de passe doit contenir au moins 8 caractères, une lettre majuscule, une lettre minuscule, un chiffre et un caractère spécial", http.StatusBadRequest)
			return
		}

		// Hashage du mot de passe
		hashedPassword, err := hashPassword(password)
		if err != nil {
			http.Error(w, "Erreur lors du hashage du mot de passe", http.StatusInternalServerError)
			return
		}

		// Création d'un nouvel utilisateur
		userID := uuid.New().String()
		_, err = db.Exec("INSERT INTO Users (id, username, email, password) VALUES (?, ?, ?, ?)", userID, username, email, hashedPassword)
		if err != nil {
			http.Error(w, "Erreur lors de l'enregistrement de l'utilisateur", http.StatusInternalServerError)
			return
		}



		// Redirection vers la page d'accueil
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	}
}

// Handler pour afficher la page de connexion et gérer la connexion
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	templateLogin := filepath.Join(projectPath, "templates", "page.login.html")
	if r.Method == "GET" {
		http.ServeFile(w, r, templateLogin)
	} else if r.Method == "POST" {
		email := r.FormValue("email")
		password := r.FormValue("password")

		var hash string
		var userID string

		// Récupérer le mot de passe hashé et l'ID de l'utilisateur à partir de l'email
		err := db.QueryRow("SELECT id, password FROM users WHERE email = ?", email).Scan(&userID, &hash)
		if err != nil {
			http.Error(w, "Utilisateur non trouvé", http.StatusUnauthorized)
			return
		}

		if checkPasswordHash(password, hash) {
			sessionID := uuid.New().String()

			// Stocker la session dans la base de données
			_, err = db.Exec("INSERT INTO Sessions (session_id, user_id, expires_at) VALUES (?, ?, ?)", sessionID, userID, time.Now().Add(24*time.Hour))
			if err != nil {
				http.Error(w, "Erreur lors de la création de la session", http.StatusInternalServerError)
				return
			}

			sessionCookie := &http.Cookie{
				Name:     "session_token",
				Value:    sessionID,
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
			}
			http.SetCookie(w, sessionCookie)

			userIDCookie := &http.Cookie{
				Name:     "userID",
				Value:    userID,
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
			}
			http.SetCookie(w, userIDCookie)

			http.Redirect(w, r, "/home", http.StatusSeeOther)
		} else {
			http.Error(w, "Mot de passe incorrect", http.StatusUnauthorized)
		}
	} else {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

func GetUserID(cookieId string) string {
	var userID string
	err := db.QueryRow("SELECT user_id FROM sessions WHERE session_id = ?", cookieId).Scan(&userID)
	if err != nil {
		return ""
	}
	return userID
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Récupérer le cookie de session
	cookie, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Error(w, "Utilisateur non authentifié", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Erreur lors de la récupération du cookie", http.StatusBadRequest)
		return
	}

	sessionID := cookie.Value

	// Supprimer la session de la base de données
	_, err = db.Exec("DELETE FROM sessions WHERE session_id = ?", sessionID)
	if err != nil {
		http.Error(w, "Erreur lors de la suppression de la session", http.StatusInternalServerError)
		return
	}

	// Invalider le cookie de session
	cookie = &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour), // Date d'expiration passée pour invalider le cookie
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	cookie = &http.Cookie{
		Name:     "userID",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour), // Date d'expiration passée pour invalider le cookie
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func validateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
