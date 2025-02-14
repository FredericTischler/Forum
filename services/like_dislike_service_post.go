package services

import (
	"database/sql"
	"errors"
)

type LikeModel struct {
	DB *sql.DB
}

func (m *LikeModel) LikePostInsert(postId, userId, action string) error {
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
		return errors.New("action invalide")
	}

	stmt := `
        INSERT INTO LikeDislikePost (post_id, user_id, like, dislike)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(post_id, user_id)
        DO UPDATE SET like = excluded.like, dislike = excluded.dislike;
    `
	_, err := m.DB.Exec(stmt, postId, userId, likeValue, dislikeValue)
	return err
}

func (m *LikeModel) VerifyAction(postId, userId string) (string, error) {
	if userId == "" {
		return "none", nil // Aucun utilisateur connecté, donc aucune action
	}

	stmt := `SELECT like, dislike FROM LikeDislikePost WHERE post_id = ? AND user_id = ?`
	var likeValue, dislikeValue int
	err := m.DB.QueryRow(stmt, postId, userId).Scan(&likeValue, &dislikeValue)
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

func (m *LikeModel) UpdateAction(postId, userId, action string) error {
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

	stmt := `UPDATE LikeDislikePost SET like = ?, dislike = ? WHERE post_id = ? AND user_id = ?`
	result, err := m.DB.Exec(stmt, likeValue, dislikeValue, postId, userId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		// Si aucune ligne n'a été mise à jour, insérez une nouvelle entrée
		return m.LikePostInsert(postId, userId, action)
	}
	return nil
}

func (m *LikeModel) CountLikesDislikes(postId int) (int, int, error) {
	row := m.DB.QueryRow(`
        SELECT 
            COALESCE(SUM(like), 0) AS like_count,
            COALESCE(SUM(dislike), 0) AS dislike_count
        FROM LikeDislikePost
        WHERE post_id = ?;
    `, postId)

	var likeCount, dislikeCount int
	err := row.Scan(&likeCount, &dislikeCount)
	if err != nil {
		return 0, 0, err
	}

	return likeCount, dislikeCount, nil
}
