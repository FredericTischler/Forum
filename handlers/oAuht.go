package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

// Configuration OAuth
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
}

// Structures pour les réponses des providers
type GithubUser struct {
	Login string `json:"login"`
	Email string `json:"email"`
}

type GoogleUser struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

var (
	githubConfig = OAuthConfig{
		ClientID:     AppConfig.GithubClientID,
		ClientSecret: AppConfig.GithubClientSecret,
		RedirectURI:  "https://localhost:8080/callback-github",
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		UserInfoURL:  "https://api.github.com/user",
	}

	googleConfig = OAuthConfig{
		ClientID:     AppConfig.GoogleClientID,
		ClientSecret: AppConfig.GoogleClientSecret,
		RedirectURI:  "https://localhost:8080/callback-google",
		AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://www.googleapis.com/oauth2/v3/userinfo",
	}
)

// Handlers GitHub
func GithubLoginHandler(w http.ResponseWriter, r *http.Request) {
	authURL := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&scope=email%%20profile",
		"https://github.com/login/oauth/authorize",
		AppConfig.GithubClientID,
		url.QueryEscape("https://localhost:8080/callback-github"))

	log.Printf("Github Client ID utilisé : %s", AppConfig.GithubClientID)
	log.Printf("URL générée : %s", authURL)

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func GithubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Début du callback GitHub")

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not received", http.StatusBadRequest)
		return
	}

	// Échange du code contre un token
	requestBody, err := json.Marshal(map[string]string{
		"client_id":     AppConfig.GithubClientID,
		"client_secret": AppConfig.GithubClientSecret,
		"code":          code,
	})
	if err != nil {
		log.Printf("Erreur lors de la préparation de la requête: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Erreur lors de la création de la requête token: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erreur lors de l'échange du code: %v", err)
		http.Error(w, "Erreur lors de l'échange du code", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		log.Printf("Erreur lors du parsing de la réponse token: %v", err)
		http.Error(w, "Erreur lors du parsing de la réponse token", http.StatusInternalServerError)
		return
	}

	// Obtention des informations utilisateur
	req, err = http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		log.Printf("Erreur lors de la création de la requête user info: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+tokenResponse.AccessToken)
	req.Header.Set("Accept", "application/json")

	userResp, err := client.Do(req)
	if err != nil {
		log.Printf("Erreur lors de la récupération des infos utilisateur: %v", err)
		http.Error(w, "Erreur lors de la récupération des infos utilisateur", http.StatusInternalServerError)
		return
	}
	defer userResp.Body.Close()

	var githubUser struct {
		Login string `json:"login"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&githubUser); err != nil {
		log.Printf("Erreur lors du parsing des infos utilisateur: %v", err)
		http.Error(w, "Erreur lors du parsing des infos utilisateur", http.StatusInternalServerError)
		return
	}

	// Si l'email n'est pas disponible dans le profil principal, faire une requête supplémentaire

	if githubUser.Email == "" {
		log.Printf("Email non trouvé dans le profil principal, tentative de récupération via l'API emails")

		emailReq, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
		if err != nil {
			log.Printf("Erreur lors de la création de la requête emails: %v", err)
			http.Error(w, "Erreur serveur", http.StatusInternalServerError)
			return
		}
		emailReq.Header.Set("Authorization", "Bearer "+tokenResponse.AccessToken)
		emailReq.Header.Set("Accept", "application/json")

		emailResp, err := client.Do(emailReq)
		if err != nil {
			log.Printf("Erreur lors de la récupération des emails: %v", err)
			http.Error(w, "Erreur lors de la récupération des emails", http.StatusInternalServerError)
			return
		}
		defer emailResp.Body.Close()

		// Lire le corps de la réponse pour le debug
		body, err := ioutil.ReadAll(emailResp.Body)
		if err != nil {
			log.Printf("Erreur lors de la lecture de la réponse email: %v", err)
			return
		}
		log.Printf("Réponse email brute: %s", string(body))

		// Structure correcte pour les emails GitHub
		type GithubEmail struct {
			Email      string `json:"email"`
			Primary    bool   `json:"primary"`
			Verified   bool   `json:"verified"`
			Visibility string `json:"visibility"`
		}

		var emails []GithubEmail
		if err := json.Unmarshal(body, &emails); err != nil {
			log.Printf("Erreur lors du parsing des emails: %v", err)
			// En cas d'erreur, on peut utiliser le login comme email de secours
			githubUser.Email = githubUser.Login + "@github.com"
			log.Printf("Utilisation de l'email de secours: %s", githubUser.Email)
		} else {
			// Chercher l'email primaire et vérifié
			for _, email := range emails {
				if email.Primary && email.Verified {
					githubUser.Email = email.Email
					log.Printf("Email primaire trouvé: %s", email.Email)
					break
				}
			}
			// Si aucun email primaire n'est trouvé, prendre le premier email vérifié
			if githubUser.Email == "" {
				for _, email := range emails {
					if email.Verified {
						githubUser.Email = email.Email
						log.Printf("Email vérifié trouvé: %s", email.Email)
						break
					}
				}
			}
		}
	}

	log.Printf("Informations utilisateur reçues: Login=%s, Email=%s", githubUser.Login, githubUser.Email)

	// Vérifier si l'utilisateur existe déjà
	var userID string
	err = db.QueryRow("SELECT id FROM Users WHERE email = ?", githubUser.Email).Scan(&userID)
	if err == sql.ErrNoRows {
		// Créer un nouvel utilisateur
		userID = uuid.New().String()
		randomPassword := uuid.New().String()
		hashedPassword, err := hashPassword(randomPassword)
		if err != nil {
			log.Printf("Erreur lors du hashage du mot de passe: %v", err)
			http.Error(w, "Erreur serveur", http.StatusInternalServerError)
			return
		}

		username := githubUser.Login
		if githubUser.Name != "" {
			username = githubUser.Name
		}

		// Vérifier si le username existe déjà
		var existingUsername string
		err = db.QueryRow("SELECT username FROM Users WHERE username = ?", username).Scan(&existingUsername)
		if err != sql.ErrNoRows {
			// Le username existe déjà, ajouter un suffixe unique
			username = fmt.Sprintf("%s_%s", username, userID[:8])
			log.Printf("Username modifié pour être unique: %s", username)
		}

		_, err = db.Exec(
			"INSERT INTO Users (id, username, email, password) VALUES (?, ?, ?, ?)",
			userID, username, githubUser.Email, hashedPassword,
		)
		if err != nil {
			log.Printf("Erreur lors de la création de l'utilisateur: %v", err)
			http.Error(w, "Erreur lors de la création de l'utilisateur", http.StatusInternalServerError)
			return
		}
		log.Printf("Nouvel utilisateur créé avec ID: %s et username: %s", userID, username)
	}

	// Créer une session
	sessionID := uuid.New().String()
	_, err = db.Exec(
		"INSERT INTO Sessions (session_id, user_id, expires_at) VALUES (?, ?, ?)",
		sessionID, userID, time.Now().Add(24*time.Hour),
	)
	if err != nil {
		log.Printf("Erreur lors de la création de la session: %v", err)
		http.Error(w, "Erreur lors de la création de la session", http.StatusInternalServerError)
		return
	}

	// Définir les cookies
	sessionCookie := &http.Cookie{
		Name:     "session_token",
		Value:    sessionID,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	}
	http.SetCookie(w, sessionCookie)

	userIDCookie := &http.Cookie{
		Name:     "userID",
		Value:    userID,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	}
	http.SetCookie(w, userIDCookie)

	log.Printf("Session créée et cookies définis. Redirection vers /home")
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

// Handlers Google
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	authURL := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&scope=email%%20profile",
		"https://accounts.google.com/o/oauth2/v2/auth",
		AppConfig.GoogleClientID,
		url.QueryEscape("https://localhost:8080/callback-google"))

	log.Printf("Google Client ID utilisé : %s", AppConfig.GoogleClientID)
	log.Printf("URL générée : %s", authURL)

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Début du callback Google")

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not received", http.StatusBadRequest)
		return
	}

	// Échange du code contre un token
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", AppConfig.GoogleClientID)
	data.Set("client_secret", AppConfig.GoogleClientSecret)
	data.Set("redirect_uri", "https://localhost:8080/callback-google")
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		log.Printf("Erreur lors de l'échange du code: %v", err)
		http.Error(w, "Erreur lors de l'échange du code", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		log.Printf("Erreur lors du parsing de la réponse token: %v", err)
		http.Error(w, "Erreur lors du parsing de la réponse token", http.StatusInternalServerError)
		return
	}

	// Obtention des informations utilisateur
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		log.Printf("Erreur lors de la création de la requête userinfo: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+tokenResponse.AccessToken)

	client := &http.Client{}
	userResp, err := client.Do(req)
	if err != nil {
		log.Printf("Erreur lors de la récupération des infos utilisateur: %v", err)
		http.Error(w, "Erreur lors de la récupération des infos utilisateur", http.StatusInternalServerError)
		return
	}
	defer userResp.Body.Close()

	var googleUser struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&googleUser); err != nil {
		log.Printf("Erreur lors du parsing des infos utilisateur: %v", err)
		http.Error(w, "Erreur lors du parsing des infos utilisateur", http.StatusInternalServerError)
		return
	}

	log.Printf("Informations utilisateur reçues: Email=%s, Name=%s", googleUser.Email, googleUser.Name)

	// Vérifier si l'utilisateur existe déjà
	var userID string
	err = db.QueryRow("SELECT id FROM Users WHERE email = ?", googleUser.Email).Scan(&userID)
	if err == sql.ErrNoRows {
		// Créer un nouvel utilisateur
		userID = uuid.New().String()
		randomPassword := uuid.New().String()
		hashedPassword, err := hashPassword(randomPassword)
		if err != nil {
			log.Printf("Erreur lors du hashage du mot de passe: %v", err)
			http.Error(w, "Erreur serveur", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(
			"INSERT INTO Users (id, username, email, password) VALUES (?, ?, ?, ?)",
			userID, googleUser.Name, googleUser.Email, hashedPassword,
		)
		if err != nil {
			log.Printf("Erreur lors de la création de l'utilisateur: %v", err)
			http.Error(w, "Erreur lors de la création de l'utilisateur", http.StatusInternalServerError)
			return
		}
		log.Printf("Nouvel utilisateur créé avec ID: %s", userID)
	} else if err != nil {
		log.Printf("Erreur lors de la recherche de l'utilisateur: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	// Créer une session
	sessionID := uuid.New().String()
	_, err = db.Exec(
		"INSERT INTO Sessions (session_id, user_id, expires_at) VALUES (?, ?, ?)",
		sessionID, userID, time.Now().Add(24*time.Hour),
	)
	if err != nil {
		log.Printf("Erreur lors de la création de la session: %v", err)
		http.Error(w, "Erreur lors de la création de la session", http.StatusInternalServerError)
		return
	}

	// Définir les cookies
	sessionCookie := &http.Cookie{
		Name:     "session_token",
		Value:    sessionID,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true, // Important pour HTTPS
		Path:     "/",
	}
	http.SetCookie(w, sessionCookie)

	userIDCookie := &http.Cookie{
		Name:     "userID",
		Value:    userID,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true, // Important pour HTTPS
		Path:     "/",
	}
	http.SetCookie(w, userIDCookie)

	log.Printf("Session créée et cookies définis. Redirection vers /home")
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

// Fonction commune pour gérer la création/connexion d'utilisateur OAuth
func handleOAuthUser(w http.ResponseWriter, r *http.Request, email, username string) {
	var userID string

	// Vérifier si l'utilisateur existe déjà
	err := db.QueryRow("SELECT id FROM Users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		// L'utilisateur n'existe pas, créer un nouveau
		userID = uuid.New().String()

		// Pour l'OAuth, on met un mot de passe aléatoire
		randomPassword := uuid.New().String()
		hashedPassword, err := hashPassword(randomPassword)
		if err != nil {
			http.Error(w, "Erreur lors du hashage du mot de passe", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(
			"INSERT INTO Users (id, username, email, password) VALUES (?, ?, ?, ?)",
			userID, username, email, hashedPassword,
		)
		if err != nil {
			http.Error(w, "Erreur lors de la création de l'utilisateur", http.StatusInternalServerError)
			return
		}
	}

	// Création de la session
	sessionID := uuid.New().String()
	_, err = db.Exec(
		"INSERT INTO Sessions (session_id, user_id, expires_at) VALUES (?, ?, ?)",
		sessionID, userID, time.Now().Add(24*time.Hour),
	)
	if err != nil {
		http.Error(w, "Erreur lors de la création de la session", http.StatusInternalServerError)
		return
	}

	// Configuration des cookies
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
}
