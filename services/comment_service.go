package services

import (
	"database/sql"
	"errors"
	"fmt"
	"forum/models"
	"strconv"
	"time"

	"github.com/google/uuid" // Assurez-vous d'importer le package UUID
)

type CommentModel struct {
	LikeModelComment *LikeModelComment
	DB               *sql.DB
}

// Insère un commentaire dans la base de données pour un post spécifique
func (m *CommentModel) CommentInsert(postId int, content, userId string) (int, error) {
	if m.DB == nil {
		return 0, errors.New("database connection is not initialized")
	}
	query := `INSERT INTO Comment (post_id, content, user_id) VALUES (?, ?, ?)`
	result, err := m.DB.Exec(query, postId, content, userId)
	if err != nil {
		return 0, fmt.Errorf("failed to insert comment: %w", err)
	}
	commentId, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}
	return int(commentId), nil
}

// Récupère les commentaires pour un post spécifique
func (m *CommentModel) GetComments(postId int, userId string) ([]models.Comment, error) {
	if m.DB == nil {
		return nil, errors.New("la connexion à la base de données n'est pas initialisée")
	}

	stmt := `SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, u.username, u.picture
             FROM Comment c
             JOIN Users u ON c.user_id = u.id
             WHERE c.post_id = ?
             ORDER BY c.created_at DESC`

	rows, err := m.DB.Query(stmt, postId)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des commentaires : %v", err)
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		var createdAt time.Time
		var commentUserId string
		var username string
		var userPicture string

		err := rows.Scan(&c.ID, &c.PostID, &commentUserId, &c.Content, &createdAt, &username, &userPicture)
		if err != nil {
			return nil, fmt.Errorf("échec de la lecture d'une ligne de commentaire : %v", err)
		}

		// Conversion de l'ID utilisateur en UUID
		c.UserID.Id, err = uuid.Parse(commentUserId)
		if err != nil {
			return nil, fmt.Errorf("ID utilisateur invalide : %v", err)
		}
		c.UserID.Username = username
		c.UserID.Picture = userPicture
		c.CreatedAt = createdAt

		// Utiliser le LikeModel pour obtenir les likes et dislikes
		likeCountComment, DislikeCountComment, err := m.LikeModelComment.CountLikesDislikesComment(c.ID)
		if err != nil {
			return nil, err
		}
		c.LikeCountComment, c.DislikeCountComment = likeCountComment, DislikeCountComment

		if userId != "" {
			UserAction, err := m.LikeModelComment.VerifyActionComment(strconv.Itoa(c.ID), userId)
			if err != nil {
				return nil, err
			}
			c.UserAction = UserAction
		} else {
			c.UserAction = ""
		}

		comments = append(comments, c)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors de l'itération des lignes : %v", err)
	}

	return comments, nil
}

// Supprime un commentaire en fonction de son ID
func (m *CommentModel) Delete(id string) error {
	if m.DB == nil {
		return errors.New("database connection is not initialized")
	}

	stmt := `DELETE FROM Comment WHERE id = ?`
	_, err := m.DB.Exec(stmt, id)
	if err != nil {
		return errors.New("failed to delete comment: " + err.Error())
	}

	return nil
}

// Met à jour le contenu d'un commentaire
func (m *CommentModel) Update(id, content string) error {
	if m.DB == nil {
		return errors.New("database connection is not initialized")
	}

	stmt := `UPDATE Comment SET content = ? WHERE id = ?`
	_, err := m.DB.Exec(stmt, content, id)
	if err != nil {
		return errors.New("failed to update comment: " + err.Error())
	}

	return nil
}

func (m *CommentModel) GetUserIdByCommentId(commentId string) (string, error) {
	if m.DB == nil {
		return "", errors.New("database connection is not initialized")
	}
	var author string
	query := `SELECT user_id FROM Comment WHERE id = ?`
	err := m.DB.QueryRow(query, commentId).Scan(&author)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("author not found")
		}
		return "", err
	}

	return author, nil
}

func (m *CommentModel) GetCommentByIdComment(commentId int) (models.Comment, error) {
	if m.DB == nil {
		return models.Comment{}, errors.New("database connection is not initialized")
	}
	var comment models.Comment
	query := `SELECT id, user_id, post_id, content, created_at FROM Comment WHERE id = ?`
	err := m.DB.QueryRow(query, commentId).Scan(&comment.ID, &comment.UserID.Id, &comment.PostID, &comment.Content, &comment.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Comment{}, errors.New("comment not found")
		}
		return models.Comment{}, err
	}

	return comment, nil
}

func (m *CommentModel) GetPostIdByCommentId(commentId string) (int, error) {
	if m.DB == nil {
		return 0, errors.New("database connection is not initialized")
	}
	var postId int
	query := `SELECT post_id FROM Comment WHERE id = ?`
	err := m.DB.QueryRow(query, commentId).Scan(&postId)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("post not found")
		}
		return 0, err
	}
	return postId, nil
}

func (m *CommentModel) GetCommentId(postId int, content, userId string) (int, error) {
	if m.DB == nil {
		return 0, errors.New("database connection is not initialized")
	}
	var commentId int
	query := `SELECT id FROM Comment WHERE post_id = ? AND content = ? AND user_id = ?`
	err := m.DB.QueryRow(query, postId, content, userId).Scan(&commentId)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("comment not found")
		}
		return 0, err
	}
	return commentId, nil
}
