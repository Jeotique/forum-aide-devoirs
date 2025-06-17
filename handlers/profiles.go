package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aide-devoir-forum/config"
	"aide-devoir-forum/database"
	"aide-devoir-forum/middleware"
	"aide-devoir-forum/models"
	"aide-devoir-forum/utils"
)

// ProfileHandler gère les profils utilisateur
type ProfileHandler struct {
	repo      *database.Repository
	config    *config.Config
	templates *template.Template
}

// NewProfileHandler crée une nouvelle instance
func NewProfileHandler(repo *database.Repository, config *config.Config, templates *template.Template) *ProfileHandler {
	return &ProfileHandler{
		repo:      repo,
		config:    config,
		templates: templates,
	}
}

// Profile affiche le profil d'un utilisateur
func (h *ProfileHandler) Profile(w http.ResponseWriter, r *http.Request) {
	// Extraire le username depuis l'URL (/profile/username)
	path := strings.TrimPrefix(r.URL.Path, "/profile/")
	username := strings.Split(path, "/")[0]

	if username == "" {
		http.Error(w, "Utilisateur non spécifié", http.StatusBadRequest)
		return
	}

	// Récupérer l'utilisateur connecté
	var currentUser *models.User
	if user, ok := r.Context().Value(middleware.UserContextKey).(*models.User); ok {
		currentUser = user
		// Mettre à jour la dernière connexion
		h.repo.UpdateLastLogin(currentUser.ID)
	}

	// Récupérer le profil demandé par username
	profileUserByUsername, err := h.repo.GetUserByUsername(username)
	if err != nil {
		h.renderError(w, "Utilisateur introuvable", "Cet utilisateur n'existe pas.", currentUser)
		return
	}

	// Récupérer le profil complet par ID
	profileUser, err := h.repo.GetUserByIDComplete(profileUserByUsername.ID)
	if err != nil {
		h.renderError(w, "Erreur", "Impossible de charger le profil.", currentUser)
		return
	}

	// Vérifier les permissions de visibilité
	canViewProfile := true
	isOwnProfile := currentUser != nil && currentUser.ID == profileUser.ID

	if !isOwnProfile && profileUser.ProfileVisibility == "private" {
		canViewProfile = false
	}

	if !canViewProfile {
		h.renderError(w, "Profil privé", "Ce profil est privé.", currentUser)
		return
	}

	// Récupérer l'activité récente
	recentActivity, err := h.repo.GetUserActivity(profileUser.ID, 10)
	if err != nil {
		recentActivity = []models.UserActivity{}
	}

	// Mettre à jour les statistiques
	h.repo.UpdateUserStats(profileUser.ID)

	data := models.ProfilePageData{
		ProfileUser:    *profileUser,
		IsOwnProfile:   isOwnProfile,
		CanViewProfile: canViewProfile,
		RecentActivity: recentActivity,
		User:           currentUser,
		Title:          fmt.Sprintf("Profil de %s", profileUser.Username),
	}

	h.renderTemplate(w, "profile.html", data)
}

// Settings affiche la page de paramètres du profil
func (h *ProfileHandler) Settings(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Récupérer le profil complet
	fullUser, err := h.repo.GetUserByIDComplete(user.ID)
	if err != nil {
		h.renderError(w, "Erreur", "Impossible de charger votre profil.", user)
		return
	}

	data := models.SettingsPageData{
		User:  fullUser,
		Title: "Paramètres du profil",
	}

	h.renderTemplate(w, "settings.html", data)
}

// UpdateProfile met à jour le profil d'un utilisateur
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok {
		h.sendJSONError(w, "Non authentifié", http.StatusUnauthorized)
		return
	}

	fmt.Printf("DEBUG UpdateProfile: Utilisateur %s (ID: %d)\n", user.Username, user.ID)

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		fmt.Printf("DEBUG UpdateProfile: Erreur ParseMultipartForm: %v\n", err)
		h.sendJSONError(w, "Données invalides", http.StatusBadRequest)
		return
	}

	// Nettoyer et valider les données
	bio := strings.TrimSpace(r.FormValue("bio"))
	location := strings.TrimSpace(r.FormValue("location"))
	visibility := r.FormValue("profile_visibility")

	fmt.Printf("DEBUG UpdateProfile: bio='%s', location='%s', visibility='%s'\n", bio, location, visibility)

	if len(bio) > 500 {
		h.sendJSONError(w, "La bio ne peut pas dépasser 500 caractères", http.StatusBadRequest)
		return
	}

	if len(location) > 100 {
		h.sendJSONError(w, "La localisation ne peut pas dépasser 100 caractères", http.StatusBadRequest)
		return
	}

	if visibility != "public" && visibility != "private" {
		visibility = "public"
	}

	fmt.Printf("DEBUG UpdateProfile: Avant appel UpdateUserProfile\n")

	// Mettre à jour en base
	err := h.repo.UpdateUserProfile(user.ID, bio, location, visibility)
	if err != nil {
		fmt.Printf("DEBUG UpdateProfile: Erreur UpdateUserProfile: %v\n", err)
		h.sendJSONError(w, "Erreur lors de la mise à jour", http.StatusInternalServerError)
		return
	}

	fmt.Printf("DEBUG UpdateProfile: Mise à jour réussie\n")

	h.sendJSONSuccess(w, "Profil mis à jour avec succès")
}

// UpdateAvatar met à jour l'avatar d'un utilisateur
func (h *ProfileHandler) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok {
		h.sendJSONError(w, "Non authentifié", http.StatusUnauthorized)
		return
	}

	// Parser le formulaire multipart
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		h.sendJSONError(w, "Fichier trop volumineux", http.StatusBadRequest)
		return
	}

	// Récupérer le fichier uploadé
	file, header, err := r.FormFile("avatar")
	if err != nil {
		h.sendJSONError(w, "Aucun fichier sélectionné", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Valider le fichier
	if err := utils.ValidateImageFile(file, header); err != nil {
		h.sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Lire le contenu du fichier
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		h.sendJSONError(w, "Erreur de lecture du fichier", http.StatusInternalServerError)
		return
	}

	// Générer un nom de fichier unique
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("avatar_%d_%d%s", user.ID, time.Now().Unix(), ext)

	// Créer le chemin complet
	uploadPath := filepath.Join("uploads", "avatars", filename)

	// Créer le dossier si nécessaire
	if err := os.MkdirAll(filepath.Dir(uploadPath), 0755); err != nil {
		h.sendJSONError(w, "Erreur de création du dossier", http.StatusInternalServerError)
		return
	}

	// Sauvegarder le fichier
	if err := os.WriteFile(uploadPath, fileBytes, 0644); err != nil {
		h.sendJSONError(w, "Erreur de sauvegarde", http.StatusInternalServerError)
		return
	}

	// Supprimer l'ancien avatar s'il existe
	oldUser, err := h.repo.GetUserByIDComplete(user.ID)
	if err == nil && oldUser.AvatarFilename != nil && *oldUser.AvatarFilename != "" {
		oldPath := filepath.Join("uploads", "avatars", *oldUser.AvatarFilename)
		os.Remove(oldPath) // Ignorer l'erreur
	}

	// Mettre à jour en base
	err = h.repo.UpdateUserAvatar(user.ID, filename)
	if err != nil {
		// Supprimer le fichier si la BDD a échoué
		os.Remove(uploadPath)
		h.sendJSONError(w, "Erreur de mise à jour en base", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message":    "Avatar mis à jour avec succès",
		"avatar_url": "/uploads/avatars/" + filename,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ServeAvatar sert les fichiers d'avatar
func (h *ProfileHandler) ServeAvatar(w http.ResponseWriter, r *http.Request) {
	// Extraire le nom du fichier depuis l'URL
	path := strings.TrimPrefix(r.URL.Path, "/uploads/avatars/")
	filename := strings.Split(path, "/")[0]

	// Sécurité : vérifier que le nom de fichier ne contient pas de chemins relatifs
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		http.NotFound(w, r)
		return
	}

	filePath := filepath.Join("uploads", "avatars", filename)

	// Vérifier que le fichier existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, filePath)
}

// === MÉTHODES UTILITAIRES ===

// renderTemplate rend un template avec gestion d'erreur
func (h *ProfileHandler) renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	if h.templates == nil {
		http.Error(w, "Templates non disponibles", http.StatusInternalServerError)
		return
	}

	if err := h.templates.ExecuteTemplate(w, templateName, data); err != nil {
		http.Error(w, "Erreur de rendu", http.StatusInternalServerError)
		return
	}
}

// renderError rend une page d'erreur
func (h *ProfileHandler) renderError(w http.ResponseWriter, title, message string, user *models.User) {
	data := map[string]interface{}{
		"Title":   title,
		"Message": message,
		"User":    user,
	}

	w.WriteHeader(http.StatusNotFound)
	h.renderTemplate(w, "error.html", data)
}

// sendJSONError envoie une réponse JSON d'erreur
func (h *ProfileHandler) sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// sendJSONSuccess envoie une réponse JSON de succès
func (h *ProfileHandler) sendJSONSuccess(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}
 