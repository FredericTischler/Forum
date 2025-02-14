package handlers

// Description: Gestion des routes pour les posts (création, modification, suppression), en intégrant les catégories.

import (
	"errors"
	"fmt"
	"forum/config"
	"forum/models"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type AppWrapper struct {
	App *config.App
}

var projectPath = config.GetProjectPath()

var ErrPostNotFound = errors.New("post not found")

const maxUploadSize = 20 << 20

// GetHome displays the home page with all posts and their categories.
func (aw AppWrapper) GetHome(w http.ResponseWriter, r *http.Request) {
	// Attempt to retrieve the "userID" cookie
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
	var notification bool

	if sessionID != "" {
		// Retrieve the username from the session ID
		username, err = aw.App.Sessions.GetUsername2(sessionID)
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}

		// verfifie si l'utilisateur a des notifications
		notification, err = aw.App.Notification.HaveNotifications(sessionID)
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}

	} else {
		username = ""
	}

	// Retrieve posts from the database using sessionID
	posts, err := aw.App.Posts.All(sessionID)
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
		"notif":    notification,
	}

	// Load the HTML template
	templatePath := filepath.Join(projectPath, "templates", "page.home.html")
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

// CreatePost displays the form to create a new post, including category selection.
func (aw AppWrapper) CreatePost(w http.ResponseWriter, r *http.Request) {
	// Retrieve all available categories to display in the form

	// Assuming you have an instance of CategoryModel in your app
	usersCookie, err := r.Cookie("userID")
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusUnauthorized, "Missing user cookie")
		return
	}

	userID := usersCookie.Value
	if userID == "" {
		aw.ErrorHandler(w, r, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	err = aw.App.Category.InitializeCategories()
	if err != nil {
		log.Fatalf("Failed to initialize categories: %v", err)
	}

	categories, err := aw.App.Category.GetAllCategory()
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Prepare data for the template
	data := map[string]interface{}{
		"categories": categories,
	}

	templatePath := filepath.Join(projectPath, "templates", "page.createpost.html")
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}
}

// StoredPost handles the submission of a new post, including selected categories.
func (aw AppWrapper) StoredPost(w http.ResponseWriter, r *http.Request) {
	// Specify the maximum file size for uploads
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// Parse the multipart form
	err := r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusBadRequest, "The uploaded file is too big. Please choose an file that's less than 20MB in size")
		return
	}

	// Retrieve form data
	title := r.PostFormValue("title")
	content := r.PostFormValue("content")

	// Retrieve selected categories (can be up to 2)
	categories := r.PostForm["categories"]
	if len(categories) == 0 {
		aw.ErrorHandler(w, r, http.StatusBadRequest, "Please select at least one category")
		return
	}
	if len(categories) > 2 {
		aw.ErrorHandler(w, r, http.StatusBadRequest, "You can select up to 2 categories")
		return
	}

	// Retrieve the session cookie
	userCookie, err := r.Cookie("session_token")
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusUnauthorized, err.Error())
		return
	}

	// Retrieve the user ID from the session
	sessionID := userCookie.Value
	userId, err := aw.App.Sessions.GetUserID(sessionID)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusUnauthorized, err.Error())
		return
	}

	// Initialize imagePath as empty
	var imagePath string

	// Retrieve the image file (if any)
	file, handler, err := r.FormFile("image")
	if err != nil {
		// If the error is http.ErrMissingFile, no image was provided
		if !errors.Is(err, http.ErrMissingFile) {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {
				aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
			}
		}(file) // Always close the file after use

		// Check the file extension
		validExtensions := map[string]bool{".jpeg": true, ".jpg": true, ".png": true, ".gif": true, ".svg": true, ".webp": true}
		ext := filepath.Ext(handler.Filename)
		if !validExtensions[ext] {
			aw.ErrorHandler(w, r, http.StatusBadRequest, "Invalid image type")
			return
		}

		// Define the path where the image will be saved
		imagePath = filepath.Join(projectPath, "static", "images_post", handler.Filename)

		// Create the file in the specified directory
		dst, err := os.Create(imagePath)
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}
		defer dst.Close() // Close the destination file after use

		// Copy the uploaded file's content to the destination file
		_, err = io.Copy(dst, file)
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Insert the post with the image (only if imagePath is not empty)
	var imageName string
	if imagePath != "" {
		imageName = handler.Filename
	}

	fmt.Println(imageName)

	// Insert the post into the database
	err = aw.App.Posts.Insert(title, content, imageName, categories, userId) // Pass the categories slice
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Redirect the user to the home page
	http.Redirect(w, r, "/home", http.StatusFound)
}

// ShowAllPost displays all posts created by the current user.
func (aw AppWrapper) ShowAllPost(w http.ResponseWriter, r *http.Request) {
	userCookie, err := r.Cookie("session_token")
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusUnauthorized, err.Error())
		return
	}

	sessionID := userCookie.Value

	// Retrieve the user ID from the session
	userId, err := aw.App.Sessions.GetUserID(sessionID)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusUnauthorized, err.Error())
		return
	}

	// Retrieve the user's posts
	posts, err := aw.App.Posts.AllPostByUser(userId)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Prepare the template
	templatePath := filepath.Join(projectPath, "templates", "post.page.html")
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Execute the template with the posts data
	err = t.Execute(w, map[string]interface{}{"posts": posts})
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
	}
}

// EditPost allows the user to edit an existing post, including its categories.
func (aw AppWrapper) EditPost(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/post/edit/"):]

	// Retrieve the user ID from the cookie
	sessionUser, err := r.Cookie("userID")
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusUnauthorized, "Missing user cookie")
		return
	}

	userID := sessionUser.Value

	// Check if the user is authorized to edit the post
	idPostUser := aw.App.Posts.GetUserPost(id)
	if userID != idPostUser {
		aw.ErrorHandler(w, r, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	// Assuming you have an instance of CategoryModel in your app
	err = aw.App.Category.InitializeCategories()
	if err != nil {
		log.Fatalf("Failed to initialize categories: %v", err)
	}

	if r.Method == http.MethodGet {
		// Retrieve the post by ID
		post, err := aw.App.Posts.Get(id)
		if errors.Is(err, ErrPostNotFound) {
			aw.ErrorHandler(w, r, http.StatusNotFound, err.Error())
			return
		} else if err != nil {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}

		// Retrieve all available categories to display in the form
		categories, err := aw.App.Category.GetAllCategory()
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}

		// Prepare data for the template, including the post and categories
		data := map[string]interface{}{
			"post":       post,
			"categories": categories,
		}

		// Define the isCategorySelected function for the template
		funcMap := template.FuncMap{
			"isCategorySelected": func(categoryName string, postCategories []models.Category) bool {
				for _, c := range postCategories {
					if c.Name == categoryName {
						return true
					}
				}
				return false
			},
		}

		// Parse the template with the function map
		templatePath := filepath.Join(projectPath, "templates", "page.setting.html")
		t, err := template.New("page.setting.html").Funcs(funcMap).ParseFiles(templatePath)
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}

		// Execute the template with the data
		err = t.Execute(w, data)
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		}
		return // Return to avoid proceeding to the POST block
	}

	if r.Method == http.MethodPost {
		err := r.ParseMultipartForm(20 << 20)
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusBadRequest, err.Error())
			return
		}

		// Retrieve form data
		title := r.PostFormValue("title")
		content := r.PostFormValue("content")

		// Retrieve selected categories
		categories := r.PostForm["categories"]
		if len(categories) == 0 {
			aw.ErrorHandler(w, r, http.StatusBadRequest, "Please select at least one category")
			return
		}
		if len(categories) > 2 {
			aw.ErrorHandler(w, r, http.StatusBadRequest, "You can select up to 2 categories")
			return
		}

		if title == "" || content == "" {
			aw.ErrorHandler(w, r, http.StatusBadRequest, "Please fill in all fields")
			return
		}

		// Initialize imageName as empty
		var imageName string

		// Retrieve the image file (if any)
		file, handler, err := r.FormFile("image")
		if err != nil {
			// If no image is provided, continue without updating the image
			if err == http.ErrMissingFile {
				// Do nothing; imageName remains empty
			} else {
				aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
				return
			}
		} else {
			defer file.Close() // Always close the file after use

			// Check the file extension
			validExtensions := map[string]bool{
				".jpeg": true, ".jpg": true, ".png": true,
				".gif": true, ".svg": true, ".webp": true,
			}
			ext := strings.ToLower(filepath.Ext(handler.Filename))
			if !validExtensions[ext] {
				aw.ErrorHandler(w, r, http.StatusBadRequest, "Invalid image type")
				return
			}

			// Define the path where the image will be saved
			imagePath := filepath.Join(projectPath, "static", "images_post", handler.Filename)

			// Create the file in the specified directory
			dst, err := os.Create(imagePath)
			if err != nil {
				aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
				return
			}
			defer dst.Close() // Close the destination file after use

			// Copy the uploaded file's content to the destination file
			_, err = io.Copy(dst, file)
			if err != nil {
				aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
				return
			}

			// Assign the image name
			imageName = handler.Filename
		}

		// Update the post with or without a new image, passing the categories
		err = aw.App.Posts.Update(id, title, content, imageName, categories)
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}

		// Redirect to the updated post
		http.Redirect(w, r, fmt.Sprintf("/post/direct/%s", id), http.StatusSeeOther)
		return
	}

	// If the HTTP method is neither GET nor POST
	aw.ErrorHandler(w, r, http.StatusMethodNotAllowed, "Method not allowed")
}

// DeletePost deletes a post and its associated image and categories.
func (aw AppWrapper) DeletePost(w http.ResponseWriter, r *http.Request) {
	// Retrieve the post ID from the URL
	id := r.URL.Path[len("/post/delete/"):]

	// Check if the ID is empty
	if id == "" {
		aw.ErrorHandler(w, r, http.StatusBadRequest, "Missing post ID")
		return
	}

	sessionUser, err := r.Cookie("userID")
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusUnauthorized, err.Error())
		return
	}

	userID := sessionUser.Value

	if userID == "" {
		aw.ErrorHandler(w, r, http.StatusUnauthorized, err.Error())
		return
	}

	idPostUser := aw.App.Posts.GetUserPost(id)

	if userID != idPostUser {
		aw.ErrorHandler(w, r, http.StatusUnauthorized, err.Error())
		return
	}

	// Retrieve the post to get the associated image
	post, err := aw.App.Posts.Get(id)
	if err != nil {
		if errors.Is(err, ErrPostNotFound) {
			aw.ErrorHandler(w, r, http.StatusNotFound, err.Error())
			return
		}
		http.Error(w, "Failed to retrieve post", http.StatusInternalServerError)
		return
	}

	// Delete the image from the file system (if it exists)
	if post.Image != nil {
		imagePath := filepath.Join(projectPath, "static", "images_post", *post.Image)
		if err := os.Remove(imagePath); err != nil {
			// Log or handle the image deletion error if necessary
			fmt.Printf("Failed to delete image: %v\n", err)
		}
	}

	// Delete the post from the database (categories will be handled by the PostModel)
	err = aw.App.Posts.Delete(id)
	if err != nil {
		http.Error(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusBadRequest, "Invalid post ID")
		return
	}
	err = aw.App.Activity.DeleteActivityByPostID(idInt)
	if err != nil {
		http.Error(w, "Failed to delete activity", http.StatusInternalServerError)
		return
	}

	// Redirect to the home page or display a success message
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

// ShowPost displays a single post along with its comments and categories.
func (aw AppWrapper) ShowPost(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/post/direct/"):]

	// Attempt to retrieve the session cookie
	sessionCookie, err := r.Cookie("session_token")
	var sessionID string
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			sessionID = ""
		} else {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		sessionID = sessionCookie.Value
	}

	var username string
	var userId string
	if sessionID != "" {
		// Retrieve the username and user ID from the session ID
		username, err = aw.App.Sessions.GetUsername(sessionID)
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}
		userId, err = aw.App.Sessions.GetUserID(sessionID)
		if err != nil {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		username = ""
		userId = ""
	}

	// Retrieve the post from the database using userId
	post, err := aw.App.Posts.GetPostUser(id, userId)
	if err != nil {
		if errors.Is(err, ErrPostNotFound) {
			aw.ErrorHandler(w, r, http.StatusNotFound, err.Error())
		} else {
			aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Retrieve comments associated with the post
	comments, err := aw.App.Comment.GetComments(post.ID, userId)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Load the template
	templatePath := filepath.Join(projectPath, "templates", "page.post.html")
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Execute the template with the post and comments
	data := map[string]interface{}{
		"userId":   userId,
		"username": username,
		"post":     post,
		"Comments": comments,
	}
	if err := t.Execute(w, data); err != nil {
		aw.ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}
}
