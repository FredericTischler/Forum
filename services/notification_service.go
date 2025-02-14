package services

import (
	"database/sql"
	"fmt"
	"forum/models"
)

type Notification struct {
	DB *sql.DB
}

func (n *Notification) AddNotification(userId string, postId int, commentId int, notifType string) error {
	var (
		ownerId string
		err     error
	)

	// Vérifier l'ownership pour un commentaire ou un post
	if commentId != 0 {
		err = n.DB.QueryRow("SELECT user_id FROM Comment WHERE id = ?", commentId).Scan(&ownerId)
		if err != nil {
			return fmt.Errorf("failed to get comment owner: %w", err)
		}
	} else {
		err = n.DB.QueryRow("SELECT user_id FROM Post WHERE id = ?", postId).Scan(&ownerId)
		if err != nil {
			return fmt.Errorf("failed to get post owner: %w", err)
		}
	}

	// Si l'utilisateur est le propriétaire, ne pas ajouter la notification
	if userId == ownerId {
		return nil
	}

	// Préparer `commentId` pour l'insertion (NULL si 0)
	var nullableCommentId interface{}
	if commentId == 0 {
		nullableCommentId = nil
	} else {
		nullableCommentId = commentId
	}

	// Déterminer le type de notification opposé
	var oppositeNotifType string
	switch notifType {
	case "like":
		oppositeNotifType = "dislike"
	case "dislike":
		oppositeNotifType = "like"
	default:
		oppositeNotifType = ""
	}

	// Supprimer la notification opposée si elle existe
	if oppositeNotifType != "" {
		queryDeleteOpposite := `
            DELETE FROM Notification
            WHERE user_id = ? AND user_id2 = ? AND post_id = ? AND comment_id IS ? AND type = ?
        `
		_, err = n.DB.Exec(queryDeleteOpposite, ownerId, userId, postId, nullableCommentId, oppositeNotifType)
		if err != nil {
			return fmt.Errorf("failed to delete opposite notification: %w", err)
		}
	}

	// Supprimer la notification du même type pour gérer les relikes/redislikes
	queryDeleteSame := `
        DELETE FROM Notification
        WHERE user_id = ? AND user_id2 = ? AND post_id = ? AND comment_id IS ? AND type = ?
    `
	result, err := n.DB.Exec(queryDeleteSame, ownerId, userId, postId, nullableCommentId, notifType)
	if err != nil {
		return fmt.Errorf("failed to delete existing notification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	// Si une notification du même type a été supprimée, ne pas réinsérer (annulation d'action)
	if rowsAffected > 0 {
		return nil
	}

	// Ajouter une nouvelle notification
	queryInsert := `
        INSERT INTO Notification (user_id, user_id2, post_id, comment_id, type, read)
        VALUES (?, ?, ?, ?, ?, ?)
    `
	_, err = n.DB.Exec(queryInsert, ownerId, userId, postId, nullableCommentId, notifType, false)
	if err != nil {
		return fmt.Errorf("failed to add notification: %w", err)
	}

	return nil
}

func (n *Notification) AddCommentNotification(commentId int) error {
	var (
		postId      int
		ownerId     string
		commenterId string
		err         error
	)

	// Récupérer l'ID du post et l'ID de l'auteur du commentaire
	queryComment := `
        SELECT post_id, user_id FROM Comment WHERE id = ?
    `
	err = n.DB.QueryRow(queryComment, commentId).Scan(&postId, &commenterId)
	if err != nil {
		return fmt.Errorf("failed to get comment details: %w", err)
	}

	// Récupérer l'ID de l'auteur du post
	queryPost := `
        SELECT user_id FROM Post WHERE id = ?
    `
	err = n.DB.QueryRow(queryPost, postId).Scan(&ownerId)
	if err != nil {
		return fmt.Errorf("failed to get post owner: %w", err)
	}

	// Afficher les variables pour le débogage
	fmt.Printf("postId: %d, ownerId: %s, commenterId: %s, commentId: %d\n", postId, ownerId, commenterId, commentId)

	// Si l'auteur du commentaire est le propriétaire du post, ne pas ajouter la notification
	if commenterId == ownerId {
		return nil
	}

	// Ajouter la notification pour le propriétaire du post
	queryInsert := `
		INSERT INTO Notification (user_id, user_id2, post_id, comment_id, type, read)
		VALUES (?,?,NULL, ?, ?, ?)
	`
	notifType := "comment"
	_, err = n.DB.Exec(queryInsert, ownerId, commenterId, commentId, notifType, false)
	if err != nil {
		return fmt.Errorf("failed to add notification: %w", err)
	}

	return nil
}

func (n *Notification) HaveNotifications(userId string) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM Notification WHERE user_id = ? AND read = 0
	`
	err := n.DB.QueryRow(query, userId).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to get notification count: %w", err)
	}

	return count > 0, nil
}

func (n *Notification) GetNotification(userId string) ([]models.Notification, error) {
	var notifications []models.Notification

	query := `
		SELECT 
			n.id, n.user_id, n.user_id2, n.post_id, n.comment_id, n.type, n.read, n.created_at,
			u1.id, u1.username, u1.email, u1.picture, u1.role, u1.created_at,
			u2.id, u2.username, u2.email, u2.picture, u2.role, u2.created_at,
			c.content, c.post_id,
			p.title
		FROM Notification n
		LEFT JOIN Users u1 ON n.user_id = u1.id
		LEFT JOIN Users u2 ON n.user_id2 = u2.id
		LEFT JOIN Comment c ON n.comment_id = c.id
		LEFT JOIN Post p ON n.post_id = p.id
		WHERE n.user_id = ? AND n.read = 0
	`

	rows, err := n.DB.Query(query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var notif models.Notification
		var user1, user2 models.User
		var postId sql.NullInt64
		var commentId sql.NullInt64
		var commentContent sql.NullString
		var commentPostId sql.NullInt64
		var postTitle sql.NullString

		err = rows.Scan(
			&notif.Id, &notif.UserId.Id, &notif.UserId2.Id, &postId, &commentId,
			&notif.Type, &notif.IsRead, &notif.CreatedAt,
			&user1.Id, &user1.Username, &user1.Email, &user1.Picture, &user1.Roles, &user1.CreatedAt,
			&user2.Id, &user2.Username, &user2.Email, &user2.Picture, &user2.Roles, &user2.CreatedAt,
			&commentContent, &commentPostId,
			&postTitle,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		// Associate users
		notif.UserId = user1
		notif.UserId2 = user2

		// Handle NULL values for Post_Id and Comment_Id
		if postId.Valid {
			notif.Post_Id = &models.Post{
				ID:    int(postId.Int64),
				Title: postTitle.String,
			}
		} else {
			notif.Post_Id = nil
		}

		if commentId.Valid {
			notif.Comment_Id = &models.Comment{
				ID:      int(commentId.Int64),
				Content: commentContent.String,
				PostID:  int(commentPostId.Int64),
			}
		} else {
			notif.Comment_Id = nil
		}

		notifications = append(notifications, notif)
	}

	return notifications, nil
}

func (n *Notification) ReadNotification(userId string, notifId int) error {
	query := `
		UPDATE Notification
		SET read = 1
		WHERE user_id = ? AND id = ?
	`
	_, err := n.DB.Exec(query, userId, notifId)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	return nil
}

func (n *Notification) GetPostId(notifId int) (int, error) {
	var postId sql.NullInt64
	var commentId sql.NullInt64

	query := `
		SELECT post_id, comment_id FROM Notification WHERE id = ?
	`
	err := n.DB.QueryRow(query, notifId).Scan(&postId, &commentId)
	if err != nil {
		return 0, fmt.Errorf("failed to get post ID and comment ID: %w", err)
	}

	if postId.Valid {
		return int(postId.Int64), nil
	}

	if commentId.Valid {
		queryComment := `
			SELECT post_id FROM Comment WHERE id = ?
		`
		err = n.DB.QueryRow(queryComment, commentId.Int64).Scan(&postId)
		if err != nil {
			return 0, fmt.Errorf("failed to get post ID from comment: %w", err)
		}

		if postId.Valid {
			return int(postId.Int64), nil
		}
	}

	return 0, fmt.Errorf("post ID not found")
}
