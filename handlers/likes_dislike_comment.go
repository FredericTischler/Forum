package handlers

import (
	"fmt"
	"net/http"
)

func (aw AppWrapper) LikeComment(w http.ResponseWriter, r *http.Request) {
	// Récupération de l'ID du comment depuis l'URL
	commentId := r.URL.Path[len("/comment/like/"):]
	postId, _ := aw.App.Comment.GetPostIdByCommentId(commentId)
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
	oldAction, err := aw.App.CommentLikes.VerifyActionComment(commentId, authorId)
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
	err = aw.handleActionChangeComment(commentId, authorId, oldAction, newAction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirection après l'action
	http.Redirect(w, r, fmt.Sprintf("/post/direct/%d", postId), http.StatusFound)
}

func (aw AppWrapper) handleActionChangeComment(commentId, authorId, oldAction, newAction string) error {
	// Validation de la nouvelle action
	validActions := map[string]bool{"like": true, "dislike": true, "none": true}
	if !validActions[newAction] {
		return fmt.Errorf("action invalide : '%s'", newAction)
	}

	// Si la nouvelle action est identique à l'ancienne, l'utilisateur annule son action
	if newAction == oldAction {
		newAction = "none"
	}

	err := aw.App.CommentLikes.LikeCommentInsert(commentId, authorId, newAction)
	// Appel de LikeCommentInsert pour insérer ou mettre à jour l'action
	return err
}
