package main

import (
	"crypto/tls"
	"database/sql"
	"forum/config"
	"forum/handlers"
	"forum/services"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	ProjectPath := config.GetProjectPath()
	dbPath := filepath.Join(ProjectPath, "app.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	handlers.LoadConfig()

	app := &config.App{
		Posts: &services.PostModel{
			DB: db,
			LikeModel: &services.LikeModel{
				DB: db,
			},
		},
		Comment: &services.CommentModel{
			DB: db,
			LikeModelComment: &services.LikeModelComment{
				DB: db,
			},
		},
		Sessions: &services.Session{
			DB: db,
		},
		Likes: &services.LikeModel{
			DB: db,
		},
		Category: &services.CategoryModel{
			DB: db,
		},
		User: &services.UserModel{
			DB: db,
		},
		CommentLikes: &services.LikeModelComment{
			DB: db,
		},
		Notification: &services.Notification{
			DB: db,
		},
		Activity: &services.Activity{
			DB: db,
		},
	}

	imagePath := filepath.Join(ProjectPath, "static", "images_post")
	imageProf := filepath.Join(ProjectPath, "static", "images_profile")
	appWrapper := &handlers.AppWrapper{App: app}
	mux := http.NewServeMux()
	mux.HandleFunc("/home", appWrapper.GetHome)
	mux.HandleFunc("/login-github", handlers.GithubLoginHandler)
	mux.HandleFunc("/callback-github", handlers.GithubCallbackHandler)
	mux.HandleFunc("/login-google", handlers.GoogleLoginHandler)
	mux.HandleFunc("/callback-google", handlers.GoogleCallbackHandler)
	mux.HandleFunc("/post/create", appWrapper.CreatePost)
	mux.HandleFunc("POST /post/create", appWrapper.StoredPost)
	mux.Handle(imagePath, http.StripPrefix(imagePath, http.FileServer(http.Dir(imagePath))))
	mux.HandleFunc("/post/", appWrapper.ShowAllPost)
	mux.HandleFunc("/post/edit/{id}", appWrapper.EditPost)
	mux.HandleFunc("/post/delete/{id}", appWrapper.DeletePost)
	mux.HandleFunc("/post/direct/{id}", appWrapper.ShowPost)
	mux.HandleFunc("/post/comment/{id}", appWrapper.HandlerCommentStore)
	mux.HandleFunc("/register", handlers.RegisterHandler)
	mux.HandleFunc("/login", handlers.LoginHandler)
	mux.HandleFunc("/logout", handlers.LogoutHandler)
	mux.HandleFunc("/post/like/{id}", appWrapper.LikePost)
	mux.HandleFunc("/post/likehome/{id}", appWrapper.LikePostHome)
	mux.HandleFunc("/post/likeprofile/{id}", appWrapper.LikeProfile)
	mux.HandleFunc("/post/likepostlike/{id}", appWrapper.LikePostLike)
	mux.HandleFunc("/post/likepostcategory/{id}", appWrapper.LikePostCategory)
	mux.HandleFunc("/profile/{username}", appWrapper.Profile)
	mux.HandleFunc("/profile/edit/{username}", appWrapper.EditProfile)
	mux.Handle(imageProf, http.StripPrefix(imageProf, http.FileServer(http.Dir(imageProf)))) // Handler pour les images de profil
	mux.HandleFunc("/category/{name}", appWrapper.GetAllPostByCat)
	mux.HandleFunc("/like/{username}", appWrapper.LikedPagePost)

	mux.HandleFunc("/comment/delete/{id}", appWrapper.DeleteComment)
	mux.HandleFunc("/comment/edit/{id}", appWrapper.EditComment)
	mux.HandleFunc("/comment/like/{id}", appWrapper.LikeComment)

	mux.HandleFunc("/notification", appWrapper.Notification)
	mux.HandleFunc("/notification/read/{id}", appWrapper.ReadNotification)
	mux.HandleFunc("/activity", appWrapper.ActivityPageHandler)

	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		}
		http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))).ServeHTTP(w, r)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		appWrapper.ErrorHandler(w, r, http.StatusNotFound, "Page non trouvée")
	})

	// Configuration TLS personnalisée
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	server := &http.Server{
		Addr:              ":8080",          //adresse du server (le port choisi est à titre d'exemple)
		Handler:           mux,              // listes des handlers
		TLSConfig:         tlsConfig,        // configuration TLS
		ReadHeaderTimeout: 10 * time.Second, // temps autorisé pour lire les headers
		ReadTimeout:       10 * time.Second, // temps maximum de lecture de la requête
		WriteTimeout:      10 * time.Second, // temps maximum d'écriture de la réponse
		IdleTimeout:       60 * time.Second, // temps maximum entre deux rêquetes
		MaxHeaderBytes:    1 << 20,          // 1 MB // maxinmum de bytes que le serveur va lire
	}

	log.Printf("Démarrage du serveur sur https://localhost%s/home", server.Addr)
	if err := server.ListenAndServeTLS("certificate.pem", "private-key.pem"); err != nil {
		log.Fatalf("Erreur du serveur: %v", err)
	}
}
