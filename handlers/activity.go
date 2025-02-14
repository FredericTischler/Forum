package handlers

import (
	"fmt"
	"forum/models"
	"html/template"
	"net/http"
	"path/filepath"
)

func (aw *AppWrapper) ActivityPageHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve user ID from cookie
	userCookie, err := r.Cookie("userID")
	if err != nil {
		http.Error(w, "Unable to retrieve user ID", http.StatusUnauthorized)
		return
	}
	userID := userCookie.Value

	var Username string

	Username, err = aw.App.Sessions.GetUsername2(userID)
	if err != nil {
		http.Error(w, "Unable to retrieve username", http.StatusUnauthorized)
		return
	}

	err = aw.App.Activity.GetPostbyUserId(userID)
	if err != nil {
		http.Error(w, "Unable to retrieve post", http.StatusInternalServerError)
		return
	}

	// Retrieve activities
	activities, err := aw.App.Activity.GetAllActivityByUser(userID)
	if err != nil {
		// Log the error to the console
		fmt.Printf("Error in GetAllActivityByUser: %v\n", err)
		// Return the error message in the HTTP response (for debugging purposes)
		http.Error(w, "Error retrieving activities: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Pass activities and username to the template
	data := struct {
		Activities []models.ActivityPage
		Username   string
	}{
		Activities: activities,
		Username:   Username,
	}

	// Load and execute the template
	templatePath := filepath.Join(projectPath, "templates", "page.activity.html")
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
	}
}
