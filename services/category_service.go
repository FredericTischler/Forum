package services

import (
	"database/sql"
	"fmt"
	"forum/models"
	"log"
	"strconv"

	"github.com/google/uuid"
)

// CategoryModel gère les opérations liées aux catégories
type CategoryModel struct {
	DB        *sql.DB
	PostModel *PostModel
}

// NewCategoryModel est un constructeur pour CategoryModel
func NewCategoryModel(db *sql.DB, postModel *PostModel) *CategoryModel {
	return &CategoryModel{
		DB:        db,
		PostModel: postModel,
	}
}

// GetAllCategory récupère toutes les catégories avec le nombre de posts associés
func (c *CategoryModel) GetAllCategory() ([]models.Category, error) {
	stmt := `
        SELECT 
            Categories.id,
            Categories.name, 
            COUNT(Catpostrel.post_id) AS post_count
        FROM 
            Categories
        LEFT JOIN 
            Catpostrel 
        ON 
            Categories.id = Catpostrel.cat_id
        GROUP BY 
            Categories.id, Categories.name
        ORDER BY 
            Categories.name;
    `

	rows, err := c.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.PostCount); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

// InitializeCategories insère des catégories prédéfinies dans la base de données
func (c *CategoryModel) InitializeCategories() error {
	predefinedCategories := []string{"Technology", "Science", "Art", "Sports", "Music"}

	for _, categoryName := range predefinedCategories {
		// Utiliser INSERT IGNORE pour éviter les doublons
		_, err := c.DB.Exec("INSERT OR IGNORE INTO Categories (name) VALUES (?)", categoryName)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetPostsByCategoryName récupère les posts associés à une catégorie donnée
func (c *CategoryModel) GetPostsByCategoryName(name string, userid string) ([]models.Post, error) {
	// Utiliser ? au lieu de $1 pour MySQL
	query := `
		SELECT p.id, p.title, p.content, p.image, p.created_at,
			   u.id, u.username, u.picture,
			   (SELECT COUNT(*) FROM LikeDislikePost WHERE post_id = p.id AND like = 1) AS like_count,
			   (SELECT COUNT(*) FROM LikeDislikePost WHERE post_id = p.id AND dislike = 1) AS dislike_count
		FROM Post p
		INNER JOIN Catpostrel cp ON p.id = cp.post_id
		INNER JOIN Categories c ON cp.cat_id = c.id
		INNER JOIN Users u ON p.user_id = u.id
		WHERE c.name = ?
		ORDER BY p.created_at DESC
	`

	rows, err := c.DB.Query(query, name)
	if err != nil {
		log.Printf("Erreur lors de l'exécution de la requête SQL: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post

	for rows.Next() {
		var post models.Post
		var userID uuid.UUID
		var username string
		var userPicture string
		var image sql.NullString
		var likeCount int
		var dislikeCount int

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&image,
			&post.CreatedAt,
			&userID,
			&username,
			&userPicture,
			&likeCount,
			&dislikeCount,
		)
		if err != nil {
			log.Printf("Erreur lors du scan des données: %v\n", err)
			return nil, err
		}

		post.UserID = models.User{
			Id:       userID,
			Username: username,
			Picture:  userPicture,
		}
		post.LikeCount = likeCount
		post.DislikeCount = dislikeCount

		if image.Valid {
			post.Image = &image.String
		} else {
			post.Image = nil
		}

		post.Category, err = c.GetCategoriesByPostID(post.ID)
		if err != nil {
			log.Printf("Erreur lors de la récupération des catégories pour le post ID %d: %v\n", post.ID, err)
			return nil, err
		}

		if userid != "" {
			if c.PostModel != nil && c.PostModel.LikeModel != nil {
				post.UserAction, err = c.PostModel.LikeModel.VerifyAction(strconv.Itoa(post.ID), userid)
				if err != nil {
					log.Printf("Erreur lors de la récupération de l'action de l'utilisateur pour le post ID %d: %v\n", post.ID, err)
					return nil, err
				}
				fmt.Println(post.UserAction)
			}
		} else {
			post.UserAction = ""
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Erreur après l'itération des lignes: %v\n", err)
		return nil, err
	}

	return posts, nil
}

// GetCategoriesByPostID récupère les catégories associées à un post spécifique
func (c *CategoryModel) GetCategoriesByPostID(postID int) ([]models.Category, error) {
	query := `
		SELECT c.id, c.name
		FROM Categories c
		INNER JOIN Catpostrel cp ON c.id = cp.cat_id
		WHERE cp.post_id = ?
	`

	rows, err := c.DB.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

// GetCategoryByName récupère une catégorie par son nom
func (c *CategoryModel) GetCategoryByName(nameCat string) (*models.Category, error) {
	query := `
		SELECT id, name
		FROM Categories
		WHERE name = ?
	`

	row := c.DB.QueryRow(query, nameCat)

	var category models.Category
	if err := row.Scan(&category.ID, &category.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Aucune catégorie trouvée avec le nom donné
		}
		return nil, err
	}

	return &category, nil
}
