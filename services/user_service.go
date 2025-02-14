package services

import (
	"database/sql"
	"fmt"
)

type UserModel struct {
	DB *sql.DB
}

func (u *UserModel) GetByUsername(username string) (string, string, string, *string, error) {
	stmt := `SELECT id, username, picture, role FROM users WHERE username = ?`
	var id, user, role string
	var picture *string // Changez cela en un pointeur pour gérer les valeurs NULL
	err := u.DB.QueryRow(stmt, username).Scan(&id, &user, &picture, &role)
	if err != nil {
		return "", "", "", nil, err
	}
	return id, user, role, picture, nil
}

func (u *UserModel) EditProfile(username, image, id string) error {
	var err error

	if image == "" {
		_, err = u.DB.Exec(`UPDATE users SET username = ? , picture = NULL WHERE id = ?`, username, id)
	} else {
		_, err = u.DB.Exec(`UPDATE users SET username = ?, picture = ? WHERE id = ?`, username, image, id)
	}

	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	return err
}

func (u *UserModel) GetUserIdByUsername(username string) (string, error) {
	stmt := `SELECT id FROM users WHERE username = ?`
	var id string
	err := u.DB.QueryRow(stmt, username).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (u *UserModel) GetAllInfoUser(id string) (string, string, *string, error) {
	stmt := `SELECT username, email, picture FROM users WHERE id = ?`
	var user, email string
	var picture *string
	err := u.DB.QueryRow(stmt, id).Scan(&user, &email, &picture)
	if err != nil {
		return "", "", nil, err
	}
	return user, email, picture, nil
}
