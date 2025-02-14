package services

import (
	"errors"
	"mime/multipart"
	"os"
	"path/filepath"
)

// Liste des formats d'images autorisés
var allowedImageFormats = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
}

// Taille maximale autorisée pour les images (20 Mo)
const maxImageSize = 20 * 1024 * 1024 // 20 MB

// UploadImage gère l'upload d'une image, vérifie le format et la taille, puis la sauvegarde sur le serveur.
func UploadImage(file multipart.File, fileHeader *multipart.FileHeader, uploadPath string) (string, error) {
	// Vérifier la taille du fichier
	if fileHeader.Size > maxImageSize {
		return "", errors.New("le fichier est trop volumineux, la taille maximale est de 20 MB")
	}

	// Vérifier le format du fichier
	ext := filepath.Ext(fileHeader.Filename)
	if !allowedImageFormats[ext] {
		return "", errors.New("format d'image non autorisé. Formats acceptés : .jpg, .jpeg, .png, .gif")
	}

	// Générer un chemin complet pour sauvegarder l'image
	newFilename := filepath.Join(uploadPath, fileHeader.Filename)

	// Créer le fichier sur le serveur
	outFile, err := os.Create(newFilename)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	// Copier le fichier uploadé dans le fichier destination
	_, err = file.Seek(0, 0) // Remettre le curseur du fichier à zéro
	if err != nil {
		return "", err
	}
	if _, err := outFile.ReadFrom(file); err != nil {
		return "", err
	}

	return newFilename, nil
}

// ValidateImageFormat vérifie si le format d'image est autorisé.
func ValidateImageFormat(filename string) error {
	ext := filepath.Ext(filename)
	if !allowedImageFormats[ext] {
		return errors.New("format d'image non autorisé")
	}
	return nil
}
