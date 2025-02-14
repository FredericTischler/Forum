package handlers

// Description : Gestion des routes pour les likes et dislikes.
// À faire :
//
//    Gérer les requêtes HTTP pour liker ou disliker un post ou un commentaire.
//    Afficher le nombre de likes et dislikes pour chaque post et commentaire.

import (
	"fmt"
	"net/http"
	"strconv"
)

func (aw AppWrapper) LikePostHome(w http.ResponseWriter, r *http.Request) {
	// Récupération de l'ID du post depuis l'URL
	postId := r.URL.Path[len("/post/likehome/"):]
	id, err := strconv.Atoi(postId)
	if err != nil || id <= 0 {
		http.Error(w, "ID du post invalide", http.StatusBadRequest)
		return
	}

	// Vérification de la session utilisateur
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	sessionId := sessionCookie.Value
	authorId, err := aw.App.Sessions.GetUserID(sessionId)
	if err != nil || authorId == "" {
		http.Error(w, "Session invalide ou utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	// Récupération de l'action précédente de l'utilisateur (like/dislike)
	oldAction, err := aw.App.Likes.VerifyAction(postId, authorId)
	if err != nil {
		http.Error(w, "Erreur lors de la vérification de l'action", http.StatusInternalServerError)
		return
	}

	// Récupération de la nouvelle action (like/dislike/none)
	newAction := r.FormValue("action")
	if newAction == "" {
		http.Error(w, "Aucune action fournie", http.StatusBadRequest)
		return
	}

	// Gestion des actions en fonction de l'état précédent
	err = aw.handleActionChange(postId, authorId, oldAction, newAction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var notificationType string

	if newAction == "like" {
		notificationType = "like"
	} else if newAction == "dislike" {
		notificationType = "dislike"
	}

	// Ajout de la notification
	err = aw.App.Notification.AddNotification(authorId, id, 0, notificationType)
	if err != nil {
		http.Error(w, "Erreur lors de l'ajout de la notification", http.StatusInternalServerError)
		return
	}

	// add Activity
	err = aw.App.Activity.LikeDislikeActivity(authorId, notificationType, id, 0)
	if err != nil {
		http.Error(w, "Erreur lors de l'ajout de l'activité", http.StatusInternalServerError)
		return
	}

	// Redirection après l'action
	http.Redirect(w, r, "/home", http.StatusFound)
}

func (aw AppWrapper) LikePost(w http.ResponseWriter, r *http.Request) {
	// Récupération de l'ID du post depuis l'URL
	postId := r.URL.Path[len("/post/like/"):]
	id, err := strconv.Atoi(postId)
	if err != nil || id <= 0 {
		http.Error(w, "ID du post invalide", http.StatusBadRequest)
		return
	}

	// Vérification de la session utilisateur
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	sessionId := sessionCookie.Value
	authorId, err := aw.App.Sessions.GetUserID(sessionId)
	if err != nil || authorId == "" {
		http.Error(w, "Session invalide ou utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	// Récupération de l'action précédente de l'utilisateur (like/dislike)
	oldAction, err := aw.App.Likes.VerifyAction(postId, authorId)
	if err != nil {
		http.Error(w, "Erreur lors de la vérification de l'action", http.StatusInternalServerError)
		return
	}

	// Récupération de la nouvelle action (like/dislike/none)
	newAction := r.FormValue("action")
	if newAction == "" {
		http.Error(w, "Aucune action fournie", http.StatusBadRequest)
		return
	}

	// Gestion des actions en fonction de l'état précédent
	err = aw.handleActionChange(postId, authorId, oldAction, newAction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var notificationType string

	if newAction == "like" {
		notificationType = "like"
	} else if newAction == "dislike" {
		notificationType = "dislike"
	}

	// Ajout de la notification
	err = aw.App.Notification.AddNotification(authorId, id, 0, notificationType)
	if err != nil {
		http.Error(w, "Erreur lors de l'ajout de la notification", http.StatusInternalServerError)
		return
	}

	err = aw.App.Activity.LikeDislikeActivity(authorId, notificationType, id, 0)
	if err != nil {
		http.Error(w, "Erreur lors de l'ajout de l'activité", http.StatusInternalServerError)
		return
	}

	// Redirection après l'action
	http.Redirect(w, r, "/post/direct/"+postId, http.StatusFound)
}

func (aw AppWrapper) LikeProfile(w http.ResponseWriter, r *http.Request) {

	// Récupération de l'ID du post depuis l'URL
	postId := r.URL.Path[len("/post/likeprofile/"):]
	id, err := strconv.Atoi(postId)
	if err != nil || id <= 0 {
		http.Error(w, "ID du post invalide", http.StatusBadRequest)
		return
	}

	// Vérification de la session utilisateur
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	sessionId := sessionCookie.Value
	authorId, err := aw.App.Sessions.GetUserID(sessionId)
	if err != nil || authorId == "" {
		http.Error(w, "Session invalide ou utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	// Récupération de l'action précédente de l'utilisateur (like/dislike)
	oldAction, err := aw.App.Likes.VerifyAction(postId, authorId)
	if err != nil {
		http.Error(w, "Erreur lors de la vérification de l'action", http.StatusInternalServerError)
		return
	}

	// Récupération de la nouvelle action (like/dislike/none)
	newAction := r.FormValue("action")
	if newAction == "" {
		http.Error(w, "Aucune action fournie", http.StatusBadRequest)
		return
	}

	// Gestion des actions en fonction de l'état précédent
	err = aw.handleActionChange(postId, authorId, oldAction, newAction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Récupération du nom d'utilisateur depuis l'URL
	usernameUrl, err := aw.App.Posts.GetUsernameByPostID(postId)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération du nom d'utilisateur", http.StatusInternalServerError)
		return
	}

	var notificationType string

	if newAction == "like" {
		notificationType = "like"
	} else if newAction == "dislike" {
		notificationType = "dislike"
	}

	// Ajout de la notification
	err = aw.App.Notification.AddNotification(authorId, id, 0, notificationType)
	if err != nil {
		http.Error(w, "Erreur lors de l'ajout de la notification", http.StatusInternalServerError)
		return
	}

	err = aw.App.Activity.LikeDislikeActivity(authorId, notificationType, id, 0)
	if err != nil {
		http.Error(w, "Erreur lors de l'ajout de l'activité", http.StatusInternalServerError)
		return
	}

	// Redirection après l'action
	http.Redirect(w, r, "/profile/"+usernameUrl, http.StatusFound)
}

func (aw AppWrapper) LikePostLike(w http.ResponseWriter, r *http.Request) {
	// Récupération de l'ID du post depuis l'URL
	postId := r.URL.Path[len("/post/likepostlike/"):]
	id, err := strconv.Atoi(postId)
	if err != nil || id <= 0 {
		http.Error(w, "ID du post invalide", http.StatusBadRequest)
		return
	}

	// Vérification de la session utilisateur
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	sessionId := sessionCookie.Value
	authorId, err := aw.App.Sessions.GetUserID(sessionId)
	if err != nil || authorId == "" {
		http.Error(w, "Session invalide ou utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	// Récupération de l'action précédente de l'utilisateur (like/dislike)
	oldAction, err := aw.App.Likes.VerifyAction(postId, authorId)
	if err != nil {
		http.Error(w, "Erreur lors de la vérification de l'action", http.StatusInternalServerError)
		return
	}

	// Récupération de la nouvelle action (like/dislike/none)
	newAction := r.FormValue("action")
	if newAction == "" {
		http.Error(w, "Aucune action fournie", http.StatusBadRequest)
		return
	}

	// Gestion des actions en fonction de l'état précédent
	err = aw.handleActionChange(postId, authorId, oldAction, newAction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	username, err := aw.App.Sessions.GetUsername2(authorId)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération du nom d'utilisateur", http.StatusInternalServerError)
		return
	}

	var notificationType string

	if newAction == "like" {
		notificationType = "like"
	} else if newAction == "dislike" {
		notificationType = "dislike"
	}

	// Ajout de la notification
	err = aw.App.Notification.AddNotification(authorId, id, 0, notificationType)
	if err != nil {
		http.Error(w, "Erreur lors de l'ajout de la notification", http.StatusInternalServerError)
		return
	}

	err = aw.App.Activity.LikeDislikeActivity(authorId, notificationType, id, 0)
	if err != nil {
		http.Error(w, "Erreur lors de l'ajout de l'activité", http.StatusInternalServerError)
		return
	}

	// Redirection après l'action
	http.Redirect(w, r, "/like/"+username, http.StatusFound)
}

func (aw AppWrapper) LikePostCategory(w http.ResponseWriter, r *http.Request) {
	// Récupération de l'ID du post depuis l'URL
	postId := r.URL.Path[len("/post/likepostcategory/"):]
	id, err := strconv.Atoi(postId)
	if err != nil || id <= 0 {
		http.Error(w, "ID du post invalide", http.StatusBadRequest)
		return
	}

	// Vérification de la session utilisateur
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	sessionId := sessionCookie.Value
	authorId, err := aw.App.Sessions.GetUserID(sessionId)
	if err != nil || authorId == "" {
		http.Error(w, "Session invalide ou utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	// Récupération de l'action précédente de l'utilisateur (like/dislike)
	oldAction, err := aw.App.Likes.VerifyAction(postId, authorId)
	if err != nil {
		http.Error(w, "Erreur lors de la vérification de l'action", http.StatusInternalServerError)
		return
	}

	// Récupération de la nouvelle action (like/dislike/none)
	newAction := r.FormValue("action")
	if newAction == "" {
		http.Error(w, "Aucune action fournie", http.StatusBadRequest)
		return
	}

	// Gestion des actions en fonction de l'état précédent
	err = aw.handleActionChange(postId, authorId, oldAction, newAction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Récupération du nom de la catégorie du post
	post, err := aw.App.Posts.GetPostByID(id)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération du post", http.StatusInternalServerError)
		return
	}

	var notificationType string

	if newAction == "like" {
		notificationType = "like"
	} else if newAction == "dislike" {
		notificationType = "dislike"
	}

	// Ajout de la notification
	err = aw.App.Notification.AddNotification(authorId, id, 0, notificationType)
	if err != nil {
		http.Error(w, "Erreur lors de l'ajout de la notification", http.StatusInternalServerError)
		return
	}

	err = aw.App.Activity.LikeDislikeActivity(authorId, notificationType, id, 0)
	if err != nil {
		http.Error(w, "Erreur lors de l'ajout de l'activité", http.StatusInternalServerError)
		return
	}

	// Vérification que le post a au moins une catégorie
	if len(post.Category) > 0 {
		categoryName := post.Category[0].Name
		// Redirection après l'action vers la catégorie du post
		http.Redirect(w, r, "/category/"+categoryName, http.StatusFound)
	} else {
		http.Error(w, "Le post n'appartient à aucune catégorie", http.StatusBadRequest)
		return
	}
}

func (aw AppWrapper) handleActionChange(postId, authorId, oldAction, newAction string) error {
	// Validation de la nouvelle action
	validActions := map[string]bool{"like": true, "dislike": true, "none": true}
	if !validActions[newAction] {
		return fmt.Errorf("action invalide : '%s'", newAction)
	}

	// Si la nouvelle action est identique à l'ancienne, l'utilisateur annule son action
	if newAction == oldAction {
		newAction = "none"
	}

	// Appel de LikePostInsert pour insérer ou mettre à jour l'action
	return aw.App.Likes.LikePostInsert(postId, authorId, newAction)
}
