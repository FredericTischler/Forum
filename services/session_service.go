package services

import (
	"database/sql"
	"fmt"
)

type Session struct {
	DB *sql.DB
}

func (m *Session) GetUserID(sessionId string) (string, error) {
	stmt := `SELECT user_id FROM sessions WHERE session_id = ?`
	var userId string

	err := m.DB.QueryRow(stmt, sessionId).Scan(&userId)

	if err != nil {
		if err == sql.ErrNoRows {

			return "", fmt.Errorf("session invalide ou non trouvée")
		}
		return "", fmt.Errorf("erreur lors de la récupération de l'ID utilisateur: %v", err)
	}

	return userId, nil
}

func (m *Session) GetUsername(sessionId string) (string, error) {
	stmt := `SELECT username FROM Users WHERE id = (SELECT user_id FROM sessions WHERE session_id = ?)`
	var username string

	err := m.DB.QueryRow(stmt, sessionId).Scan(&username)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("aucun utilisateur trouvé pour l'ID de session %s", sessionId)
		}
		return "", fmt.Errorf("erreur lors de la récupération du nom d'utilisateur: %v", err)
	}

	return username, nil
}

func (m *Session) GetUsername2(cookieId string) (string, error) {
	stmt := `SELECT username FROM Users WHERE id = ?`
	var username string

	err := m.DB.QueryRow(stmt, cookieId).Scan(&username)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("aucun utilisateur trouvé pour l'ID de session %s", cookieId)
		}
		return "", fmt.Errorf("erreur lors de la récupération du nom d'utilisateur: %v", err)
	}

	return username, nil
}
