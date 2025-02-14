package services

import (
	"database/sql"
	"errors"
	"fmt"
	"forum/models"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PostModel struct {
	DB        *sql.DB
	LikeModel *LikeModel
}

var ErrPostNotFound = errors.New("post not found")

// Insert inserts a new post along with its categories into the database.
func (m *PostModel) Insert(title, content, image string, categories []string, userId string) error {
	var stmt string
	var res sql.Result
	var err error

	// Insert the post into the Post table
	if image == "" {
		stmt = `INSERT INTO Post (user_id, title, content, created_at)
		        VALUES(?, ?, ?, datetime('now'))`
		res, err = m.DB.Exec(stmt, userId, title, content)
	} else {
		stmt = `INSERT INTO Post (user_id, title, content, image, created_at)
		        VALUES(?, ?, ?, ?, datetime('now'))`
		res, err = m.DB.Exec(stmt, userId, title, content, image)
	}
	if err != nil {
		return err
	}

	// Get the last inserted post ID
	postID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// Insert categories into the Categories table if they don't exist
	for _, catName := range categories {
		var catID int
		// Check if the category already exists
		err = m.DB.QueryRow("SELECT id FROM Categories WHERE name = ?", catName).Scan(&catID)
		if err != nil {
			if err == sql.ErrNoRows {
				// Insert new category
				res, err = m.DB.Exec("INSERT INTO Categories (name) VALUES (?)", catName)
				if err != nil {
					return err
				}
				catID64, err := res.LastInsertId()
				if err != nil {
					return err
				}
				catID = int(catID64)
			} else {
				return err
			}
		}

		// Insert into Catpostrel table
		_, err = m.DB.Exec("INSERT INTO Catpostrel (cat_id, post_id) VALUES (?, ?)", catID, postID)
		if err != nil {
			return err
		}
	}

	return nil
}

// All retrieves all posts along with their categories.
func (m *PostModel) All(userId string) ([]models.Post, error) {
	stmt := `SELECT 
                p.id, 
                p.title, 
                p.content, 
                p.image, 
                p.created_at,
                u.id AS user_id, 
                u.username, 
                u.picture,
                GROUP_CONCAT(c.name, ',') AS categories
             FROM Post p
             JOIN users u ON p.user_id = u.id
             LEFT JOIN Catpostrel cp ON p.id = cp.post_id
             LEFT JOIN Categories c ON cp.cat_id = c.id
             GROUP BY p.id
             ORDER BY p.id DESC`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []models.Post{}
	for rows.Next() {
		var p models.Post
		var image sql.NullString
		var userPicture sql.NullString
		var userIdStr string
		var createdAt time.Time
		var categoriesStr sql.NullString

		err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Content,
			&image,
			&createdAt,
			&userIdStr,
			&p.UserID.Username,
			&userPicture,
			&categoriesStr, // Scan the concatenated categories
		)
		if err != nil {
			return nil, err
		}

		p.UserID.Id, err = uuid.Parse(userIdStr)
		if err != nil {
			return nil, err
		}

		if image.Valid {
			p.Image = &image.String
		} else {
			p.Image = nil
		}

		if userPicture.Valid {
			p.UserID.Picture = userPicture.String
		} else {
			p.UserID.Picture = "default.jpg"
		}

		p.CreatedAt = createdAt

		// Parse the concatenated categories into the Category slice
		if categoriesStr.Valid && categoriesStr.String != "" {
			categoryNames := strings.Split(categoriesStr.String, ",")
			for _, name := range categoryNames {
				p.Category = append(p.Category, models.Category{Name: strings.TrimSpace(name)})
			}
		} else {
			p.Category = []models.Category{}
		}

		// Use the injected LikeModel
		likeCount, dislikeCount, err := m.LikeModel.CountLikesDislikes(p.ID)
		if err != nil {
			return nil, err
		}
		p.LikeCount = likeCount
		p.DislikeCount = dislikeCount

		if userId != "" {
			userAction, err := m.LikeModel.VerifyAction(strconv.Itoa(p.ID), userId)
			if err != nil {
				return nil, err
			}
			p.UserAction = userAction
		} else {
			p.UserAction = ""
		}

		posts = append(posts, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// getCategoriesByPostID retrieves categories associated with a given post ID.
func (m *PostModel) getCategoriesByPostID(postID int) ([]models.Category, error) {
	stmt := `SELECT c.id, c.name
	         FROM Categories c
	         JOIN Catpostrel cp ON c.id = cp.cat_id
	         WHERE cp.post_id = ?`

	rows, err := m.DB.Query(stmt, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// AllPostByUser retrieves all posts by a specific user along with their categories.
func (m *PostModel) AllPostByUser(userid string) ([]models.Post, error) {
	stmt := `SELECT p.id, p.title, p.content, p.image, p.created_at,
	                u.id AS user_id, u.username, u.picture
	         FROM Post p
	         JOIN Users u ON p.user_id = u.id
	         WHERE p.user_id = ?
	         ORDER BY p.id DESC`

	rows, err := m.DB.Query(stmt, userid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []models.Post{}
	for rows.Next() {
		var p models.Post
		var image sql.NullString
		var userPicture sql.NullString
		var userIdStr string
		var createdAt time.Time

		err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Content,
			&image,
			&createdAt,
			&userIdStr,
			&p.UserID.Username,
			&userPicture,
		)
		if err != nil {
			return nil, err
		}

		p.UserID.Id, err = uuid.Parse(userIdStr)
		if err != nil {
			return nil, err
		}

		if image.Valid {
			p.Image = &image.String
		} else {
			p.Image = nil
		}

		if userPicture.Valid {
			p.UserID.Picture = userPicture.String
		} else {
			p.UserID.Picture = "default.jpg"
		}

		p.CreatedAt = createdAt

		// Retrieve categories for the post
		p.Category, err = m.getCategoriesByPostID(p.ID)
		if err != nil {
			return nil, err
		}

		posts = append(posts, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// AllPostByUserProfile retrieves all posts by a specific user profile along with their categories.
func (m *PostModel) AllPostByUserProfile(userid string, currentUserID string, sessionuserdID string) ([]models.Post, error) {
	stmt := `SELECT p.id, p.title, p.content, p.image, p.created_at,
	                u.id AS user_id, u.username, u.picture
	         FROM Post p
	         JOIN users u ON p.user_id = u.id
	         WHERE p.user_id = ?
	         ORDER BY p.id DESC`

	rows, err := m.DB.Query(stmt, userid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []models.Post{}
	for rows.Next() {
		var p models.Post
		var image sql.NullString
		var userPicture sql.NullString
		var userIdStr string
		var createdAt time.Time

		err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Content,
			&image,
			&createdAt,
			&userIdStr,
			&p.UserID.Username,
			&userPicture,
		)
		if err != nil {
			return nil, err
		}

		p.UserID.Id, err = uuid.Parse(userIdStr)
		if err != nil {
			return nil, err
		}

		if image.Valid {
			p.Image = &image.String
		} else {
			p.Image = nil
		}

		if userPicture.Valid {
			p.UserID.Picture = userPicture.String
		} else {
			p.UserID.Picture = "default.jpg"
		}

		p.CreatedAt = createdAt

		// Retrieve categories for the post
		p.Category, err = m.getCategoriesByPostID(p.ID)
		if err != nil {
			return nil, err
		}

		// Use the LikeModel to get likes and dislikes
		likeCount, dislikeCount, err := m.LikeModel.CountLikesDislikes(p.ID)
		if err != nil {
			return nil, err
		}
		p.LikeCount = likeCount
		p.DislikeCount = dislikeCount

		if sessionuserdID != "" {
			userAction, err := m.LikeModel.VerifyAction(strconv.Itoa(p.ID), sessionuserdID)
			if err != nil {
				return nil, err
			}
			p.UserAction = userAction
		} else {
			p.UserAction = ""
		}

		posts = append(posts, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// Get retrieves a single post by ID along with its categories.
func (pm *PostModel) Get(id string) (*models.Post, error) {
	post := &models.Post{}
	query := `SELECT p.id, p.title, p.content, p.image, p.created_at,
	                 u.id, u.username, u.picture
	          FROM Post p
	          JOIN Users u ON p.user_id = u.id
	          WHERE p.id = ?`

	// Variables for scanning
	var userIdStr string
	var image sql.NullString

	err := pm.DB.QueryRow(query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&image,
		&post.CreatedAt,
		&userIdStr,
		&post.UserID.Username,
		&post.UserID.Picture,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPostNotFound
		}
		fmt.Printf("Error retrieving post: %v\n", err)
		return nil, err
	}

	post.UserID.Id, err = uuid.Parse(userIdStr)
	if err != nil {
		return nil, err
	}

	if image.Valid {
		post.Image = &image.String
	} else {
		post.Image = nil
	}

	// Retrieve categories for the post
	post.Category, err = pm.getCategoriesByPostID(post.ID)
	if err != nil {
		return nil, err
	}

	return post, nil
}

// GetPostUser retrieves a single post by ID along with user action and categories.
func (pm *PostModel) GetPostUser(postId string, userId string) (*models.Post, error) {
	post := &models.Post{}
	likeModel := &LikeModel{DB: pm.DB}
	if userId != "" {
		userAction, err := likeModel.VerifyAction(postId, userId)
		if err != nil {
			return nil, fmt.Errorf("error getting user action on post: %v", err)
		}
		post.UserAction = userAction
	} else {
		post.UserAction = "none"
	}

	post, err := pm.Get(postId)
	if err != nil {
		return nil, err
	}

	// Get likes and dislikes
	likeCount, dislikeCount, err := pm.LikeModel.CountLikesDislikes(post.ID)
	if err != nil {
		return nil, err
	}
	post.LikeCount = likeCount
	post.DislikeCount = dislikeCount

	// Determine user action on this post
	if userId != "" {
		userAction, err := pm.LikeModel.VerifyAction(strconv.Itoa(post.ID), userId)
		if err != nil {
			return nil, err
		}
		post.UserAction = userAction
	} else {
		post.UserAction = ""
	}

	return post, nil
}

// Update updates a post's title, content, image, and categories.
func (pm *PostModel) Update(id string, title string, content string, image string, categories []string) error {
	fmt.Println("Updating Post with id: ", id)
	var err error

	// Update the Post table
	if image == "" {
		_, err = pm.DB.Exec("UPDATE Post SET title = ?, content = ?, image = NULL WHERE id = ?", title, content, id)
	} else {
		_, err = pm.DB.Exec("UPDATE Post SET title = ?, content = ?, image = ? WHERE id = ?", title, content, image, id)
	}
	if err != nil {
		return err
	}

	// Delete existing categories for this post
	_, err = pm.DB.Exec("DELETE FROM Catpostrel WHERE post_id = ?", id)
	if err != nil {
		return err
	}

	// Re-insert categories
	for _, catName := range categories {
		var catID int
		// Check if the category exists
		err = pm.DB.QueryRow("SELECT id FROM Categories WHERE name = ?", catName).Scan(&catID)
		if err != nil {
			if err == sql.ErrNoRows {
				// Insert new category
				res, err := pm.DB.Exec("INSERT INTO Categories (name) VALUES (?)", catName)
				if err != nil {
					return err
				}
				catID64, err := res.LastInsertId()
				if err != nil {
					return err
				}
				catID = int(catID64)
			} else {
				return err
			}
		}

		// Insert into Catpostrel table
		_, err = pm.DB.Exec("INSERT INTO Catpostrel (cat_id, post_id) VALUES (?, ?)", catID, id)
		if err != nil {
			return err
		}
	}

	return nil
}

// Delete removes a post and its associated category relationships.
func (pm *PostModel) Delete(id string) error {
	// Delete from Catpostrel table first due to foreign key constraints
	_, err := pm.DB.Exec("DELETE FROM Catpostrel WHERE post_id = ?", id)
	if err != nil {
		return err
	}

	// Delete from Post table
	_, err = pm.DB.Exec("DELETE FROM Post WHERE id = ?", id)
	return err
}

// GetUserPost retrieves the user ID associated with a post.
func (pm *PostModel) GetUserPost(postId string) string {
	var id string
	query := "SELECT user_id FROM Post WHERE id = ?"
	err := pm.DB.QueryRow(query, postId).Scan(&id)
	if err != nil {
		return ""
	}
	return id
}

// GetLikedPost retrieves all posts that a user has liked.
func (m *PostModel) GetLikedPost(userId string) ([]models.Post, error) {
	stmt := `SELECT 
				p.id, 
				p.title, 
				p.content, 
				p.image, 
				p.created_at,
				u.id AS user_id, 
				u.username, 
				u.picture,
				GROUP_CONCAT(c.name, ',') AS categories
			FROM 
				Post p
			JOIN 
				LikeDislikePost l ON p.id = l.post_id
			JOIN 
				users u ON p.user_id = u.id
			LEFT JOIN 
				Catpostrel cp ON p.id = cp.post_id
			LEFT JOIN 
				Categories c ON cp.cat_id = c.id
			WHERE 
				l.user_id = ? AND l.like = 1
			GROUP BY 
				p.id
			ORDER BY 
				p.id DESC`

	rows, err := m.DB.Query(stmt, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []models.Post{}
	for rows.Next() {
		var p models.Post
		var image sql.NullString
		var userPicture sql.NullString
		var userIdStr string
		var createdAt time.Time
		var categoriesStr sql.NullString

		err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Content,
			&image,
			&createdAt,
			&userIdStr,
			&p.UserID.Username,
			&userPicture,
			&categoriesStr,
		)
		if err != nil {
			return nil, err
		}

		p.UserID.Id, err = uuid.Parse(userIdStr)
		if err != nil {
			return nil, err
		}

		if image.Valid {
			p.Image = &image.String
		} else {
			p.Image = nil
		}

		if userPicture.Valid {
			p.UserID.Picture = userPicture.String
		} else {
			p.UserID.Picture = "default.jpg"
		}

		p.CreatedAt = createdAt

		if categoriesStr.Valid && categoriesStr.String != "" {
			categoryNames := strings.Split(categoriesStr.String, ",")
			for _, name := range categoryNames {
				p.Category = append(p.Category, models.Category{Name: strings.TrimSpace(name)})
			}
		} else {
			p.Category = []models.Category{}
		}

		likeCount, dislikeCount, err := m.LikeModel.CountLikesDislikes(p.ID)
		if err != nil {
			return nil, err
		}
		p.LikeCount = likeCount
		p.DislikeCount = dislikeCount

		userAction, err := m.LikeModel.VerifyAction(strconv.Itoa(p.ID), userId)
		if err != nil {
			return nil, err
		}
		p.UserAction = userAction

		posts = append(posts, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// GetPostByID retrieves a post by ID along with its categories.
func (m *PostModel) GetPostByID(id int) (*models.Post, error) {
	post := &models.Post{}
	stmt := `SELECT p.id, p.title, p.content, p.image, p.created_at,
					u.id AS user_id, u.username, u.picture,
					GROUP_CONCAT(c.name, ',') AS categories
			 FROM Post p
			 JOIN Users u ON p.user_id = u.id
			 LEFT JOIN Catpostrel cp ON p.id = cp.post_id
			 LEFT JOIN Categories c ON cp.cat_id = c.id
			 WHERE p.id = ?
			 GROUP BY p.id`

	var image sql.NullString
	var userPicture sql.NullString
	var userIdStr string
	var createdAt time.Time
	var categoriesStr sql.NullString

	err := m.DB.QueryRow(stmt, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&image,
		&createdAt,
		&userIdStr,
		&post.UserID.Username,
		&userPicture,
		&categoriesStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	post.UserID.Id, err = uuid.Parse(userIdStr)
	if err != nil {
		return nil, err
	}

	if image.Valid {
		post.Image = &image.String
	} else {
		post.Image = nil
	}

	if userPicture.Valid {
		post.UserID.Picture = userPicture.String
	} else {
		post.UserID.Picture = "default.jpg"
	}

	post.CreatedAt = createdAt

	if categoriesStr.Valid && categoriesStr.String != "" {
		categoryNames := strings.Split(categoriesStr.String, ",")
		for _, name := range categoryNames {
			post.Category = append(post.Category, models.Category{Name: strings.TrimSpace(name)})
		}
	} else {
		post.Category = []models.Category{}
	}

	return post, nil
}

func (m *PostModel) GetUsernameByPostID(postId string) (string, error) {
	var username string
	query := "SELECT u.username FROM Post p JOIN Users u ON p.user_id = u.id WHERE p.id = ?"
	err := m.DB.QueryRow(query, postId).Scan(&username)
	if err != nil {
		return "", err
	}
	return username, nil
}
