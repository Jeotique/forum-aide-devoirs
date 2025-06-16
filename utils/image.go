package utils

import (
	"crypto/rand"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Configuration pour les images
const (
	MaxFileSize    = 10 << 20 // 10MB
	MaxImagesCount = 5        // Maximum 5 images par post
	UploadPath     = "./uploads/posts"
	ThumbnailPath  = "./uploads/thumbnails"
)

// Types MIME autorisés
var AllowedMimeTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/jpg":  ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// ImageInfo contient les informations d'une image uploadée
type ImageInfo struct {
	Filename     string
	OriginalName string
	ContentType  string
	SizeBytes    int
	Width        int
	Height       int
}

// ValidateImageFile valide un fichier image
func ValidateImageFile(file multipart.File, header *multipart.FileHeader) error {
	// Vérifier la taille
	if header.Size > MaxFileSize {
		return fmt.Errorf("fichier trop volumineux (max %d MB)", MaxFileSize/(1<<20))
	}

	// Vérifier le type MIME
	contentType := header.Header.Get("Content-Type")
	if _, ok := AllowedMimeTypes[contentType]; !ok {
		return fmt.Errorf("type de fichier non autorisé: %s", contentType)
	}

	// Vérifier l'extension du fichier
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExt := false
	for _, allowedExt := range AllowedMimeTypes {
		if ext == allowedExt {
			validExt = true
			break
		}
	}
	if !validExt {
		return fmt.Errorf("extension de fichier non autorisée: %s", ext)
	}

	return nil
}

// GenerateUniqueFilename génère un nom de fichier unique (max 100 caractères)
func GenerateUniqueFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().Unix()

	// Générer un ID aléatoire plus court
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomID := fmt.Sprintf("%x", randomBytes)

	// Format: timestamp_randomID.ext (environ 20-25 caractères max)
	return fmt.Sprintf("%d_%s%s", timestamp, randomID, ext)
}

// SaveImageFile sauvegarde un fichier image sur le disque
func SaveImageFile(file multipart.File, header *multipart.FileHeader) (*ImageInfo, error) {
	// Valider le fichier
	if err := ValidateImageFile(file, header); err != nil {
		return nil, err
	}

	// Générer un nom de fichier unique
	filename := GenerateUniqueFilename(header.Filename)
	filePath := filepath.Join(UploadPath, filename)

	// Créer le dossier si nécessaire
	if err := os.MkdirAll(UploadPath, 0755); err != nil {
		return nil, fmt.Errorf("impossible de créer le dossier: %v", err)
	}

	// Créer le fichier de destination
	destFile, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("impossible de créer le fichier: %v", err)
	}
	defer destFile.Close()

	// Copier le contenu
	file.Seek(0, 0) // Remettre au début
	size, err := io.Copy(destFile, file)
	if err != nil {
		os.Remove(filePath) // Nettoyer en cas d'erreur
		return nil, fmt.Errorf("erreur lors de la copie: %v", err)
	}

	// Obtenir les dimensions de l'image
	width, height, err := getImageDimensions(filePath)
	if err != nil {
		// Ne pas échouer si on ne peut pas obtenir les dimensions
		width, height = 0, 0
	}

	return &ImageInfo{
		Filename:     filename,
		OriginalName: header.Filename,
		ContentType:  header.Header.Get("Content-Type"),
		SizeBytes:    int(size),
		Width:        width,
		Height:       height,
	}, nil
}

// getImageDimensions récupère les dimensions d'une image
func getImageDimensions(filePath string) (int, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	// Décoder l'image pour obtenir les dimensions
	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	return img.Width, img.Height, nil
}

// DeleteImageFile supprime un fichier image du disque
func DeleteImageFile(filename string) error {
	filePath := filepath.Join(UploadPath, filename)
	return os.Remove(filePath)
}

// CreateThumbnail crée une miniature d'une image (optionnel pour plus tard)
func CreateThumbnail(srcPath string, thumbPath string, maxWidth, maxHeight int) error {
	// Ouvrir l'image source
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Décoder l'image
	srcImg, format, err := image.Decode(srcFile)
	if err != nil {
		return err
	}

	// Créer le dossier thumbnail si nécessaire
	if err := os.MkdirAll(ThumbnailPath, 0755); err != nil {
		return err
	}

	// Créer le fichier thumbnail
	thumbFile, err := os.Create(thumbPath)
	if err != nil {
		return err
	}
	defer thumbFile.Close()

	// Pour cette version simple, on copie juste l'image originale
	// Une vraie implémentation utiliserait une bibliothèque de redimensionnement
	srcFile.Seek(0, 0)
	_, err = io.Copy(thumbFile, srcFile)

	// Encoder selon le format original
	if format == "jpeg" || format == "jpg" {
		return jpeg.Encode(thumbFile, srcImg, &jpeg.Options{Quality: 80})
	} else if format == "png" {
		return png.Encode(thumbFile, srcImg)
	}

	return err
}

// GetImageURL retourne l'URL publique d'une image
func GetImageURL(filename string) string {
	return "/uploads/posts/" + filename
}
