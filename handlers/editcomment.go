package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
)

func (aw AppWrapper) EditComment(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/comment/edit/"):]

	sessionUser, err := r.Cookie("userID")
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusUnauthorized, err.Error())
		http.Error(w, "Missing user cookie", http.StatusUnauthorized)
		return
	}

	userID := sessionUser.Value
	//author, err := aw.App.Sessions.GetUsername2(userID)
	userIdComment, err := aw.App.Comment.GetUserIdByCommentId(id)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	if userID != userIdComment {
		aw.ErrorHandler(w, r, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	if r.Method == http.MethodGet {
		id, err := strconv.Atoi(id)
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusBadRequest, "Invalid comment ID")
			return
		}
		comment, err := aw.App.Comment.GetCommentByIdComment(id)
		if errors.Is(err, ErrPostNotFound) {
			aw.ErrorHandler(w, r, http.StatusNotFound, err.Error())
			return
		}
		templatePath := filepath.Join(projectPath, "templates", "page.commentsetting.html")

		t, err := template.ParseFiles(templatePath)
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}

		err = t.Execute(w, map[string]interface{}{"comment": comment})
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		}
	}

	if r.Method == http.MethodPost {
		// Récupérer les données du formulaire
		content := r.PostFormValue("content")

		if content == "" {
			aw.ErrorHandler(w, r, http.StatusBadRequest, "Please fill in all fields")
			return
		}
		err = aw.App.Comment.Update(id, content)
		if err != nil {
			http.Error(w, "Unable to update comment, please try again later", http.StatusInternalServerError)
			return
		}
		id, _ := aw.App.Comment.GetPostIdByCommentId(id)
		// Rediriger vers la liste des posts ou afficher un message de succès
		http.Redirect(w, r, fmt.Sprintf("/post/direct/%d", id), http.StatusSeeOther)
	}
}
