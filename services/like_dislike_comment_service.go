package services

import (
	"database/sql"
	"errors"
	"fmt"
)

type LikeModelComment struct {
	DB *sql.DB
}

func (m *LikeModelComment) LikeCommentInsert(comment_id, userId, action string) error {
	var likeValue, dislikeValue int
	switch action {
	case "like":
		likeValue = 1
		dislikeValue = 0
	case "dislike":
		likeValue = 0
		dislikeValue = 1
	case "none":
		likeValue = 0
		dislikeValue = 0
	default:
		err := errors.New("action invalide")
		fmt.Println(err)
		return err
	}

	stmt := `
		INSERT INTO LikeDislikeComment (user_id, comment_id, like, dislike, created_at)
		VALUES (?, ?, ?, ?, Datetime('now'))
		ON CONFLICT(comment_id, user_id)
		DO UPDATE SET like = excluded.like, dislike = excluded.dislike;
	`
	_, err := m.DB.Exec(stmt, userId, comment_id, likeValue, dislikeValue)
	return err
}

func (m *LikeModelComment) VerifyActionComment(commentId, userId string) (string, error) {
	stmt := `SELECT like, dislike FROM LikeDislikeComment WHERE comment_id = ? AND user_id = ?`
	var likeValue, dislikeValue int
	err := m.DB.QueryRow(stmt, commentId, userId).Scan(&likeValue, &dislikeValue)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // Aucun like ou dislike trouvé
		}
		return "", err
	}
	if likeValue == 1 {
		return "like", nil
	} else if dislikeValue == 1 {
		return "dislike", nil
	} else {
		return "", nil
	}
}

func (m *LikeModelComment) UpdateActionComment(commentId, userId, action string) error {
	var likeValue, dislikeValue int
	switch action {
	case "like":
		likeValue = 1
		dislikeValue = 0
	case "dislike":
		likeValue = 0
		dislikeValue = 1
	case "none":
		likeValue = 0
		dislikeValue = 0
	default:
		return errors.New("action update invalide")
	}

	stmt := `UPDATE LikeDislikeComment SET like = ?, dislike = ? WHERE comment_id = ? AND user_id = ?`
	result, err := m.DB.Exec(stmt, likeValue, dislikeValue, commentId, userId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		// Si aucune ligne n'a été mise à jour, insérez une nouvelle entrée
		return m.LikeCommentInsert(commentId, userId, action)
	}
	return nil
}

func (m *LikeModelComment) CountLikesDislikesComment(commentId int) (int, int, error) {
	row := m.DB.QueryRow(`
        SELECT 
            COALESCE(SUM(like), 0) AS like_count,
            COALESCE(SUM(dislike), 0) AS dislike_count
        FROM LikeDislikeComment
        WHERE comment_id = ?;
    `, commentId)

	var likeCount, dislikeCount int
	err := row.Scan(&likeCount, &dislikeCount)
	if err != nil {
		return 0, 0, err
	}

	return likeCount, dislikeCount, nil
}
