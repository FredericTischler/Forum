package config

//Description : Fichier pour gérer la configuration globale de l'application.
//À faire :
//
//    Charger les variables d'environnement (par exemple, clés OAuth, configuration de la base de données).
//    Définir des constantes ou des paramètres globaux utilisés dans tout le projet.
import (
	"forum/services"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	Posts        *services.PostModel
	Comment      *services.CommentModel
	Sessions     *services.Session
	Likes        *services.LikeModel
	Category     *services.CategoryModel
	User         *services.UserModel
	CommentLikes *services.LikeModelComment
	Notification *services.Notification
	Activity     *services.Activity
}

// GetProjectPath retourne le chemin du répertoire racine du projet
func GetProjectPath() string {
	path, _ := os.Getwd()
	return path
}
