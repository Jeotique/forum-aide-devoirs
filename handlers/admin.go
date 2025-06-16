package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"aide-devoir-forum/config"
	"aide-devoir-forum/database"
	"aide-devoir-forum/middleware"
	"aide-devoir-forum/models"
	"aide-devoir-forum/utils"
)

type AdminHandler struct {
	repo      *database.Repository
	config    *config.Config
	templates *template.Template
}

func NewAdminHandler(repo *database.Repository, cfg *config.Config, tmpl *template.Template) *AdminHandler {
	return &AdminHandler{
		repo:      repo,
		config:    cfg,
		templates: tmpl,
	}
}

// GET /admin
func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil || !user.IsAdmin() {
		http.Error(w, "Accès refusé", http.StatusForbidden)
		return
	}

	// Récupérer tous les utilisateurs
	users, err := h.repo.GetAllUsers()
	if err != nil {
		users = []models.User{}
	}

	// Récupérer les catégories
	categories, err := h.repo.GetCategories()
	if err != nil {
		categories = []models.Category{}
	}

	// Récupérer les logs de modération
	logs, err := h.repo.GetModerationLogs(100)
	if err != nil {
		logs = []models.ModerationLog{}
	}

	// Récupérer les statistiques
	stats, err := h.repo.GetAdminStats()
	if err != nil {
		stats = models.AdminStats{}
	}

	data := models.AdminPageData{
		Users:      users,
		Categories: categories,
		Logs:       logs,
		Stats:      stats,
		User:       user,
		Title:      "Administration",
	}

	if h.templates != nil {
		utils.RenderTemplate(w, h.templates, "admin.html", data)
	} else {
		var usersHTML string
		for _, u := range users {
			banStatus := ""
			if u.IsBanned {
				banStatus = `<span class="status banned">Banni</span>`
			} else {
				banStatus = `<span class="status active">Actif</span>`
			}

			usersHTML += `<div class="user-item">
				<div class="user-info">
					<strong>` + u.Username + `</strong>
					<span class="role-badge role-` + strconv.Itoa(u.RoleID) + `">` + u.RoleName + `</span>
					` + banStatus + `
				</div>
				<div class="user-actions">
					<button onclick="banUser(` + strconv.Itoa(u.ID) + `)" class="btn btn-warning">Bannir</button>
					<button onclick="promoteUser(` + strconv.Itoa(u.ID) + `)" class="btn btn-info">Promouvoir</button>
				</div>
			</div>`
		}

		content := `
			<h1>Administration</h1>
			<div class="admin-section">
				<h2>Gestion des utilisateurs</h2>
				<div class="users-list">` + usersHTML + `</div>
			</div>
			<script>
				function banUser(userId) {
					const reason = prompt("Raison du bannissement :");
					if (reason) {
						fetch('/admin/ban', {
							method: 'POST',
							headers: {'Content-Type': 'application/x-www-form-urlencoded'},
							body: 'user_id=' + userId + '&reason=' + encodeURIComponent(reason)
						}).then(() => location.reload());
					}
				}
				function promoteUser(userId) {
					const newRole = prompt("Nouveau rôle (2=Prof, 3=Modérateur, 4=Admin) :");
					if (newRole) {
						fetch('/admin/promote', {
							method: 'POST',
							headers: {'Content-Type': 'application/x-www-form-urlencoded'},
							body: 'user_id=' + userId + '&role_id=' + newRole
						}).then(() => location.reload());
					}
				}
			</script>
		`

		utils.RenderSimplePage(w, "Administration", content)
	}
}

// POST /admin/ban
func (h *AdminHandler) BanUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	if user == nil || !user.CanModerate() {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Accès refusé"})
		return
	}

	userIDStr := r.FormValue("user_id")
	reason := r.FormValue("reason")

	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID utilisateur invalide"})
		return
	}

	if reason == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Raison requise"})
		return
	}

	// Vérifier qu'on ne bannit pas un admin
	targetUser, err := h.repo.GetUserByID(userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Utilisateur non trouvé"})
		return
	}

	if targetUser.IsAdmin() && !user.IsAdmin() {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Impossible de bannir un administrateur"})
		return
	}

	// Bannir l'utilisateur
	err = h.repo.BanUser(userID, reason)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur lors du bannissement"})
		return
	}

	// Logger l'action
	h.repo.CreateModerationLog(user.ID, "ban", "user", userID, reason)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// POST /admin/promote
func (h *AdminHandler) PromoteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	if user == nil || !user.CanPromoteUsers() {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Accès refusé"})
		return
	}

	userIDStr := r.FormValue("user_id")
	roleIDStr := r.FormValue("role_id")

	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID utilisateur invalide"})
		return
	}

	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil || roleID < 1 || roleID > 4 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Rôle invalide"})
		return
	}

	// Promouvoir l'utilisateur
	err = h.repo.PromoteUser(userID, roleID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur lors de la promotion"})
		return
	}

	// Logger l'action
	h.repo.CreateModerationLog(user.ID, "promote", "user", userID, "Rôle changé à "+roleIDStr)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// POST /admin/delete-post
func (h *AdminHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	if user == nil || !user.CanModerate() {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Accès refusé"})
		return
	}

	postIDStr := r.FormValue("post_id")
	reason := r.FormValue("reason")

	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID de post invalide"})
		return
	}

	// Supprimer le post
	err = h.repo.DeletePost(postID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur lors de la suppression"})
		return
	}

	// Logger l'action
	h.repo.CreateModerationLog(user.ID, "delete", "post", postID, reason)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// POST /admin/delete-comment
func (h *AdminHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	if user == nil || !user.CanModerate() {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Accès refusé"})
		return
	}

	commentIDStr := r.FormValue("comment_id")
	reason := r.FormValue("reason")

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil || commentID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID de commentaire invalide"})
		return
	}

	// Supprimer le commentaire
	err = h.repo.DeleteComment(commentID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur lors de la suppression"})
		return
	}

	// Logger l'action
	h.repo.CreateModerationLog(user.ID, "delete", "comment", commentID, reason)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// POST /admin/mark-solution
func (h *AdminHandler) MarkSolution(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentification requise"})
		return
	}

	commentIDStr := r.FormValue("comment_id")
	postIDStr := r.FormValue("post_id")

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil || commentID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID de commentaire invalide"})
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID de post invalide"})
		return
	}

	// Vérifier si l'utilisateur peut marquer comme solution
	// (doit être l'auteur du post ou un modérateur)
	postAuthorID, err := h.repo.GetPostAuthorID(postID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Post non trouvé"})
		return
	}

	if user.ID != postAuthorID && !user.CanModerate() {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Seul l'auteur du post peut marquer une solution"})
		return
	}

	// Marquer le commentaire comme solution
	err = h.repo.MarkCommentAsSolution(commentID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur lors du marquage"})
		return
	}

	// Marquer le post comme résolu
	err = h.repo.MarkPostAsSolved(postID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur lors du marquage du post"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// UnbanUser débannit un utilisateur
func (h *AdminHandler) UnbanUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	if user == nil || !user.CanModerate() {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Accès refusé"})
		return
	}

	userIDStr := r.FormValue("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID utilisateur invalide"})
		return
	}

	// Débannir l'utilisateur
	err = h.repo.UnbanUser(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur lors du débannissement"})
		return
	}

	// Logger l'action
	h.repo.CreateModerationLog(user.ID, "unban", "user", userID, "Utilisateur débanni")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// POST /admin/categories
func (h *AdminHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	if user == nil || !user.IsAdmin() {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Accès refusé"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
		Icon        string `json:"icon"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Données invalides"})
		return
	}

	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Le nom est requis"})
		return
	}

	err := h.repo.CreateCategory(req.Name, req.Description, req.Color, req.Icon)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur lors de la création"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// PUT /admin/categories/{id}
func (h *AdminHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	if user == nil || !user.IsAdmin() {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Accès refusé"})
		return
	}

	// Extraire l'ID de l'URL
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID manquant"})
		return
	}

	id, err := strconv.Atoi(parts[3])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID invalide"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
		Icon        string `json:"icon"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Données invalides"})
		return
	}

	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Le nom est requis"})
		return
	}

	err = h.repo.UpdateCategory(id, req.Name, req.Description, req.Color, req.Icon)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur lors de la modification"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// DELETE /admin/categories/{id}
func (h *AdminHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	if user == nil || !user.IsAdmin() {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Accès refusé"})
		return
	}

	// Extraire l'ID de l'URL
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID manquant"})
		return
	}

	id, err := strconv.Atoi(parts[3])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID invalide"})
		return
	}

	err = h.repo.DeleteCategory(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
