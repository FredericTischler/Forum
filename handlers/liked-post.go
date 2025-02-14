package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
)

func (aw AppWrapper) LikedPagePost(w http.ResponseWriter, r *http.Request) {
	userCookie, err := r.Cookie("userID")
	var sessionID string
	if err != nil {
		// If the cookie doesn't exist, sessionID remains an empty string
		sessionID = ""
	} else {
		// The cookie exists; safely access userCookie.Value
		sessionID = userCookie.Value
	}

	var username string

	if sessionID != "" {
		// Retrieve the username from the session ID
		username, err = aw.App.Sessions.GetUsername2(sessionID)
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		username = ""
	}

	// Retrieve posts from the database using sessionID
	posts, err := aw.App.Posts.GetLikedPost(sessionID)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	category, err := aw.App.Category.GetAllCategory()
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Prepare data for the template
	data := map[string]interface{}{
		"posts":    posts,
		"username": username,
		"category": category,
	}

	// Load the HTML template
	templatePath := filepath.Join(projectPath, "templates", "page.likepost.html")
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Execute the template with the data
	err = t.Execute(w, data)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
	}
}
