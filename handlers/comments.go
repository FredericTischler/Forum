package handlers

//Description : Gestion des routes pour les commentaires (ajout, suppression).
//À faire :
//
//    Gérer les requêtes HTTP pour ajouter et supprimer des commentaires.
//    Vérifier que seuls les utilisateurs enregistrés peuvent commenter.

import (
	"fmt"
	"net/http"
	"strconv"
)

func (aw AppWrapper) HandlerCommentStore(w http.ResponseWriter, r *http.Request) {
	// Vérifier si la méthode est POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extraire l'ID du post de l'URL
	idStr := r.URL.Path[len("/post/comment/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Parse le formulaire pour obtenir le contenu
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	content := r.PostFormValue("content")
	if content == "" {
		http.Error(w, "Please fill in all fields", http.StatusBadRequest)
		return
	}

	author, err := r.Cookie("userID")
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	sessionId := author.Value

	// Declare commentID
	var commentID int

	// Insertion du commentaire dans la base de données avec un postId en int
	commentID, err = aw.App.Comment.CommentInsert(id, content, sessionId)
	if err != nil {
		http.Error(w, "Unable to submit comment, please try again later", http.StatusInternalServerError)
		return
	}

	fmt.Println(commentID)

	// Add notification using the comment ID
	err = aw.App.Notification.AddCommentNotification(commentID)
	if err != nil {
		http.Error(w, "Unable to add notification, please try again later", http.StatusInternalServerError)
		return
	}
	// Add activity
	err = aw.App.Activity.CreateActivity(sessionId, "comment", id, commentID)
	if err != nil {
		http.Error(w, "Unable to add activity, please try again later", http.StatusInternalServerError)
		return
	}

	// Redirection après insertion
	http.Redirect(w, r, fmt.Sprintf("/post/direct/%d", id), http.StatusSeeOther)
}

func (aw AppWrapper) DeleteComment(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/comment/delete/"):]
	sessionUser, err := r.Cookie("userID")
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusUnauthorized, err.Error())
		http.Error(w, "Missing user cookie", http.StatusUnauthorized)
		return
	}

	userID := sessionUser.Value
	authorIdComment, err := aw.App.Comment.GetUserIdByCommentId(idStr)
	if err != nil {
		http.Error(w, "Unable to retrieve author ID", http.StatusInternalServerError)
		return
	}
	if userID != authorIdComment {
		aw.ErrorHandler(w, r, http.StatusUnauthorized, "Unauthorized access")
		return
	}
	postId, err := aw.App.Comment.GetPostIdByCommentId(idStr)
	if err != nil {
		http.Error(w, "Unable to retrieve post ID", http.StatusInternalServerError)
		return
	}

	err = aw.App.Comment.Delete(idStr)
	if err != nil {
		http.Error(w, "Unable to delete comment, please try again later", http.StatusInternalServerError)
		return
	}

	// Delete activity
	commentID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}
	
	err = aw.App.Activity.DeleteActivity(userID, "comment", postId, commentID)
	if err != nil {
		http.Error(w, "Unable to delete activity, please try again later", http.StatusInternalServerError)
		return
	}
	
	// Rediriger vers la liste des posts ou afficher un message de succès
	http.Redirect(w, r, fmt.Sprintf("/post/direct/%d", postId), http.StatusSeeOther)
}
