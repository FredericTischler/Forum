package services

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hash un mot de passe en utilisant bcrypt.
func HashPassword(password string) (string, error) {
	// Utiliser bcrypt pour générer un hash sécurisé avec un coût de 14
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword compare un mot de passe fourni avec son hash stocké.
func CheckPassword(hashedPassword, providedPassword string) bool {
	// Comparer le hash avec le mot de passe fourni
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword))
	return err == nil // Si l'erreur est nulle, le mot de passe correspond
}
