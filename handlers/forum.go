package handlers

import (
	"encoding/json"
	"fmt"
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

type ForumHandler struct {
	repo      *database.Repository
	config    *config.Config
	templates *template.Template
}

func NewForumHandler(repo *database.Repository, cfg *config.Config, tmpl *template.Template) *ForumHandler {
	return &ForumHandler{
		repo:      repo,
		config:    cfg,
		templates: tmpl,
	}
}

// GET /
func (h *ForumHandler) Home(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())

	// Récupérer les catégories
	categories, err := h.repo.GetCategories()
	if err != nil {
		categories = []models.Category{}
	}

	// Récupérer les posts récents
	recentPosts, err := h.repo.GetRecentPosts(10)
	if err != nil {
		recentPosts = []models.Post{}
	}

	data := models.HomePageData{
		Categories:  categories,
		RecentPosts: recentPosts,
		User:        user,
		Title:       "Accueil",
	}

	if h.templates != nil {
		utils.RenderTemplate(w, h.templates, "home.html", data)
	} else {
		// Fallback HTML simple
		var categoriesHTML string
		for _, cat := range categories {
			categoriesHTML += `<div class="category-card">
				<h3><i class="` + cat.Icon + `"></i> ` + cat.Name + `</h3>
				<p>` + cat.Description + `</p>
				<span class="post-count">` + strconv.Itoa(cat.PostCount) + ` posts</span>
			</div>`
		}

		var postsHTML string
		for _, post := range recentPosts {
			postsHTML += `<div class="post-preview">
				<h4><a href="/post/` + strconv.Itoa(post.ID) + `">` + post.Title + `</a></h4>
				<p>` + utils.TruncateText(post.Content, 150) + `</p>
				<small>Par ` + post.Username + ` dans ` + post.CategoryName + `</small>
			</div>`
		}

		content := `
			<h2>Catégories</h2>
			<div class="categories-grid">` + categoriesHTML + `</div>
			<h2>Posts récents</h2>
			<div class="posts-list">` + postsHTML + `</div>
		`

		utils.RenderSimplePage(w, "Forum d'aide aux devoirs", content)
	}
}

// GET /category/{id}
func (h *ForumHandler) Category(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())

	// Extraire l'ID de la catégorie
	path := r.URL.Path
	categoryID, err := utils.ExtractIDFromPath(path, "/category/")
	if err != nil {
		http.Error(w, "ID de catégorie invalide", http.StatusBadRequest)
		return
	}

	// Récupérer la catégorie
	category, err := h.repo.GetCategory(categoryID)
	if err != nil {
		http.Error(w, "Catégorie non trouvée", http.StatusNotFound)
		return
	}

	// Récupérer les posts de la catégorie
	posts, err := h.repo.GetPostsByCategory(categoryID)
	if err != nil {
		posts = []models.Post{}
	}

	data := models.CategoryPageData{
		Category: *category,
		Posts:    posts,
		User:     user,
		Title:    category.Name,
	}

	if h.templates != nil {
		utils.RenderTemplate(w, h.templates, "category.html", data)
	} else {
		var postsHTML string
		for _, post := range posts {
			badges := ""
			if post.IsSolved {
				badges += `<span class="badge solved">Résolu</span>`
			}
			if post.IsPinned {
				badges += `<span class="badge pinned">Épinglé</span>`
			}
			if post.IsLocked {
				badges += `<span class="badge locked">Verrouillé</span>`
			}

			postsHTML += `<div class="post-item">
				<h3><a href="/post/` + strconv.Itoa(post.ID) + `">` + post.Title + `</a></h3>
				<p>` + utils.TruncateText(post.Content, 200) + `</p>
				<div class="post-meta">
					<span>Par ` + post.Username + `</span>
					<span>` + utils.FormatTime(post.CreatedAt) + `</span>
					<span>👀 ` + strconv.Itoa(post.ViewsCount) + `</span>
					<span>👍 ` + strconv.Itoa(post.LikesCount) + `</span>
					` + badges + `
				</div>
			</div>`
		}

		content := `
			<h1><i class="` + category.Icon + `"></i> ` + category.Name + `</h1>
			<p>` + category.Description + `</p>
			<div class="posts-list">` + postsHTML + `</div>
		`

		utils.RenderSimplePage(w, category.Name, content)
	}
}

// GET /post/{id}
func (h *ForumHandler) Post(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())

	// Extraire l'ID du post
	path := r.URL.Path
	postID, err := utils.ExtractIDFromPath(path, "/post/")
	if err != nil {
		http.Error(w, "ID de post invalide", http.StatusBadRequest)
		return
	}

	// Récupérer le post
	post, err := h.repo.GetPost(postID, user)
	if err != nil {
		http.Error(w, "Post non trouvé", http.StatusNotFound)
		return
	}

	// Vérifier si l'utilisateur peut voir ce post (en particulier pour les posts archivés)
	userID := 0
	userRoleID := 0
	if user != nil {
		userID = user.ID
		userRoleID = user.RoleID
	}

	if !post.CanBeViewedBy(userID, userRoleID) {
		http.Error(w, "Post non trouvé", http.StatusNotFound)
		return
	}

	// Incrémenter les vues
	h.repo.IncrementPostViews(postID)

	// Récupérer le paramètre de tri des commentaires
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "newest" // par défaut
	}

	// Récupérer les commentaires avec tri
	comments, err := h.repo.GetCommentsWithSort(postID, user, sortBy)
	if err != nil {
		comments = []models.Comment{}
	}

	data := models.PostPageData{
		Post:        *post,
		Comments:    comments,
		User:        user,
		Title:       post.Title,
		CurrentSort: sortBy,
		AvailableSorts: []models.SortOption{
			{Value: "newest", Label: "Plus récents"},
			{Value: "oldest", Label: "Plus anciens"},
			{Value: "most_liked", Label: "Les plus likés"},
			{Value: "solutions_first", Label: "Solutions d'abord"},
		},
	}

	if h.templates != nil {
		utils.RenderTemplate(w, h.templates, "post.html", data)
	} else {
		// Affichage simple du post
		var tagsHTML string
		for _, tag := range post.Tags {
			tagsHTML += `<span class="tag">` + tag.Name + `</span>`
		}

		var commentsHTML string
		for _, comment := range comments {
			solutionBadge := ""
			if comment.IsSolution {
				solutionBadge = `<span class="badge solution">Solution</span>`
			}

			commentsHTML += `<div class="comment">
				<div class="comment-header">
					<strong>` + comment.Username + `</strong>
					<span class="role-badge role-` + strings.ToLower(comment.UserRole) + `">` + comment.UserRole + `</span>
					<span>` + utils.FormatTime(comment.CreatedAt) + `</span>
					` + solutionBadge + `
				</div>
				<div class="comment-content">` + comment.Content + `</div>
				<div class="comment-votes">
					<span>👍 ` + strconv.Itoa(comment.LikesCount) + `</span>
					<span>👎 ` + strconv.Itoa(comment.DislikesCount) + `</span>
				</div>
			</div>`
		}

		content := `
			<div class="post-detail">
				<h1>` + post.Title + `</h1>
				<div class="post-meta">
					<span>Par ` + post.Username + `</span>
					<span class="role-badge role-` + strings.ToLower(post.UserRole) + `">` + post.UserRole + `</span>
					<span>` + utils.FormatTime(post.CreatedAt) + `</span>
					<span>👀 ` + strconv.Itoa(post.ViewsCount) + `</span>
					<span>👍 ` + strconv.Itoa(post.LikesCount) + `</span>
					<span>👎 ` + strconv.Itoa(post.DislikesCount) + `</span>
				</div>
				<div class="post-content">` + post.Content + `</div>
				<div class="post-tags">` + tagsHTML + `</div>
			</div>
			<div class="comments-section">
				<h3>Commentaires</h3>
				` + commentsHTML + `
			</div>
		`

		utils.RenderSimplePage(w, post.Title, content)
	}
}

// GET /create-post
func (h *ForumHandler) CreatePostPage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Récupérer les catégories pour le formulaire
	categories, err := h.repo.GetCategories()
	if err != nil {
		categories = []models.Category{}
	}

	if h.templates != nil {
		data := map[string]interface{}{
			"Categories": categories,
			"User":       user,
			"Title":      "Créer un post",
		}
		utils.RenderTemplate(w, h.templates, "create-post.html", data)
	} else {
		var categoriesOptions string
		for _, cat := range categories {
			categoriesOptions += `<option value="` + strconv.Itoa(cat.ID) + `">` + cat.Name + `</option>`
		}

		content := `
			<form method="POST" action="/create-post" class="form-container">
				<div class="form-group">
					<label for="title">Titre :</label>
					<input type="text" id="title" name="title" required minlength="5" maxlength="255">
				</div>
				<div class="form-group">
					<label for="category_id">Catégorie :</label>
					<select id="category_id" name="category_id" required>
						<option value="">Choisir une catégorie</option>
						` + categoriesOptions + `
					</select>
				</div>
				<div class="form-group">
					<label for="content">Contenu :</label>
					<textarea id="content" name="content" required minlength="20" rows="10"></textarea>
				</div>
				<div class="form-group">
					<label for="tags">Tags (séparés par des virgules) :</label>
					<input type="text" id="tags" name="tags" placeholder="mathématiques, algèbre, équations">
				</div>
				<button type="submit" class="btn btn-primary">Publier</button>
			</form>
		`

		utils.RenderSimplePage(w, "Créer un post", content)
	}
}

// POST /create-post
func (h *ForumHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/create-post?error=method", http.StatusSeeOther)
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parser le formulaire multipart pour les fichiers
	err := r.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		http.Redirect(w, r, "/create-post?error=upload", http.StatusSeeOther)
		return
	}

	title := utils.SanitizeInput(r.FormValue("title"))
	content := utils.SanitizeInput(r.FormValue("content"))
	categoryIDStr := r.FormValue("category_id")
	tagsStr := r.FormValue("tags")

	// Validation
	if len(title) < 5 || len(title) > 255 {
		http.Redirect(w, r, "/create-post?error=title", http.StatusSeeOther)
		return
	}

	if len(content) < 20 {
		http.Redirect(w, r, "/create-post?error=content", http.StatusSeeOther)
		return
	}

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil || categoryID <= 0 {
		http.Redirect(w, r, "/create-post?error=category", http.StatusSeeOther)
		return
	}

	// Créer le post
	postID, err := h.repo.CreatePost(title, content, user.ID, categoryID)
	if err != nil {
		http.Redirect(w, r, "/create-post?error=create", http.StatusSeeOther)
		return
	}

	// Traiter les images uploadées
	files := r.MultipartForm.File["images"]
	if len(files) > utils.MaxImagesCount {
		// Nettoyer et retourner erreur
		h.repo.DeletePost(int(postID))
		http.Redirect(w, r, "/create-post?error=too_many_images", http.StatusSeeOther)
		return
	}

	var uploadErrors []string
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			uploadErrors = append(uploadErrors, fmt.Sprintf("Erreur ouverture %s", fileHeader.Filename))
			continue
		}
		defer file.Close()

		// Sauvegarder l'image
		imageInfo, err := utils.SaveImageFile(file, fileHeader)
		if err != nil {
			uploadErrors = append(uploadErrors, fmt.Sprintf("Erreur upload %s: %v", fileHeader.Filename, err))
			continue
		}

		// Enregistrer en base de données
		image := &models.Image{
			Filename:     imageInfo.Filename,
			OriginalName: imageInfo.OriginalName,
			ContentType:  imageInfo.ContentType,
			SizeBytes:    imageInfo.SizeBytes,
			Width:        imageInfo.Width,
			Height:       imageInfo.Height,
			PostID:       utils.IntPtr(int(postID)),
			UserID:       user.ID,
		}

		err = h.repo.CreateImage(image)
		if err != nil {
			// Supprimer le fichier physique si erreur DB
			utils.DeleteImageFile(imageInfo.Filename)
			uploadErrors = append(uploadErrors, fmt.Sprintf("Erreur DB pour %s", fileHeader.Filename))
		}
	}

	// Ajouter les tags
	if tagsStr != "" {
		tags := utils.ParseTags(tagsStr)
		for _, tag := range tags {
			h.repo.AddTagToPost(int(postID), tag)
		}
	}

	// Rediriger vers le post créé avec succès
	successURL := "/post/" + strconv.FormatInt(postID, 10) + "?success=created"
	if len(uploadErrors) > 0 {
		successURL += "&upload_warnings=1"
	}
	http.Redirect(w, r, successURL, http.StatusSeeOther)
}

// POST /comment
func (h *ForumHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentification requise"})
		return
	}

	// Parser le formulaire multipart pour les fichiers
	err := r.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur lors du parsing du formulaire"})
		return
	}

	postIDStr := r.FormValue("post_id")
	content := utils.SanitizeInput(r.FormValue("content"))
	parentIDStr := r.FormValue("parent_id")

	// Validation
	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID de post invalide"})
		return
	}

	// Vérifier que le post peut recevoir des commentaires
	post, err := h.repo.GetPost(postID, user)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Post non trouvé"})
		return
	}

	if !post.CanReceiveComments() {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ce post ne peut plus recevoir de commentaires"})
		return
	}

	if len(content) < 5 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Le commentaire doit faire au moins 5 caractères"})
		return
	}

	var parentID *int
	if parentIDStr != "" {
		if pid, err := strconv.Atoi(parentIDStr); err == nil && pid > 0 {
			parentID = &pid
		}
	}

	// Créer le commentaire et récupérer son ID
	commentID, err := h.repo.CreateCommentWithID(postID, content, user.ID, parentID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur lors de la création du commentaire"})
		return
	}

	// Traiter les images uploadées (max 3 pour les commentaires)
	files := r.MultipartForm.File["images"]
	if len(files) > 3 { // Limite réduite pour les commentaires
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Maximum 3 images autorisées pour un commentaire"})
		return
	}

	var uploadErrors []string
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			uploadErrors = append(uploadErrors, fmt.Sprintf("Erreur ouverture %s", fileHeader.Filename))
			continue
		}
		defer file.Close()

		// Sauvegarder l'image
		imageInfo, err := utils.SaveImageFile(file, fileHeader)
		if err != nil {
			uploadErrors = append(uploadErrors, fmt.Sprintf("Erreur upload %s: %v", fileHeader.Filename, err))
			continue
		}

		// Enregistrer en base de données
		image := &models.Image{
			Filename:     imageInfo.Filename,
			OriginalName: imageInfo.OriginalName,
			ContentType:  imageInfo.ContentType,
			SizeBytes:    imageInfo.SizeBytes,
			Width:        imageInfo.Width,
			Height:       imageInfo.Height,
			CommentID:    utils.IntPtr(int(commentID)),
			UserID:       user.ID,
		}

		err = h.repo.CreateImage(image)
		if err != nil {
			// Supprimer le fichier physique si erreur DB
			utils.DeleteImageFile(imageInfo.Filename)
			uploadErrors = append(uploadErrors, fmt.Sprintf("Erreur DB pour %s", fileHeader.Filename))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{"status": "success"}
	if len(uploadErrors) > 0 {
		response["upload_warnings"] = uploadErrors
	}
	json.NewEncoder(w).Encode(response)
}

// POST /vote
func (h *ForumHandler) Vote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentification requise"})
		return
	}

	voteType := r.FormValue("type")
	target := r.FormValue("target")
	targetIDStr := r.FormValue("target_id")

	// Validation
	if voteType != "like" && voteType != "dislike" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Type de vote invalide"})
		return
	}

	if target != "post" && target != "comment" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Cible de vote invalide"})
		return
	}

	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil || targetID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID de cible invalide"})
		return
	}

	// Appliquer le vote
	if target == "post" {
		err = h.repo.VotePost(targetID, user.ID, voteType)
	} else {
		err = h.repo.VoteComment(targetID, user.ID, voteType)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur lors du vote"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// GET /search
func (h *ForumHandler) Search(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())

	query := utils.GetStringParam(r, "q")
	categoryID, _ := utils.GetIntParam(r, "category")

	var posts []models.Post
	var searchInfo map[string]interface{}

	if query != "" {
		var err error
		posts, err = h.repo.SearchPosts(query, categoryID, 50)
		if err != nil {
			posts = []models.Post{}
		}

		// Analyser la requête pour les informations de recherche
		titleQuery, tags := h.parseSearchQuery(query)
		searchInfo = map[string]interface{}{
			"Query":      query,
			"TitleQuery": titleQuery,
			"Tags":       tags,
			"HasTags":    len(tags) > 0,
			"HasTitle":   titleQuery != "",
		}
	}

	// Récupérer les catégories pour le formulaire
	categories, _ := h.repo.GetCategories()

	// Récupérer les tags populaires pour les suggestions
	popularTags, _ := h.repo.GetPopularTags(10)

	if h.templates != nil {
		data := map[string]interface{}{
			"SearchInfo":  searchInfo,
			"Posts":       posts,
			"Categories":  categories,
			"PopularTags": popularTags,
			"User":        user,
			"Title":       "Recherche",
			"CategoryID":  categoryID,
		}
		utils.RenderTemplate(w, h.templates, "search.html", data)
	} else {
		var postsHTML string
		for _, post := range posts {
			var tagsHTML string
			for _, tag := range post.Tags {
				tagsHTML += `<span class="tag">` + tag.Name + `</span>`
			}

			postsHTML += `<div class="post-item">
				<h3><a href="/post/` + strconv.Itoa(post.ID) + `">` + post.Title + `</a></h3>
				<p>` + utils.TruncateText(post.Content, 200) + `</p>
				<div class="post-meta">
					<small>Par ` + post.Username + ` dans ` + post.CategoryName + `</small>
					<div class="post-tags">` + tagsHTML + `</div>
				</div>
			</div>`
		}

		var popularTagsHTML string
		for _, tag := range popularTags {
			popularTagsHTML += `<span class="tag-suggestion" onclick="addTag('` + tag.Name + `')">#` + tag.Name + `</span>`
		}

		content := `
			<div class="search-container">
				<form method="GET" action="/search" class="search-form">
					<div class="search-input-container">
						<input type="text" name="q" value="` + query + `" placeholder="Rechercher par titre ou #tag...">
						<button type="submit"><i class="fas fa-search"></i></button>
					</div>
					<div class="popular-tags">
						<strong>Tags populaires :</strong>
						` + popularTagsHTML + `
					</div>
				</form>
				<div class="search-results">
					<h2>Résultats de recherche (` + strconv.Itoa(len(posts)) + `)</h2>
					` + postsHTML + `
				</div>
			</div>
		`

		utils.RenderSimplePage(w, "Recherche", content)
	}
}

// API /search-suggestions - Retourne des suggestions de recherche en JSON
func (h *ForumHandler) SearchSuggestions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if len(query) < 2 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{})
		return
	}

	suggestions, err := h.repo.SearchSuggestions(query, 10)
	if err != nil {
		suggestions = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}

// Fonction utilitaire pour parser les requêtes de recherche
func (h *ForumHandler) parseSearchQuery(query string) (string, []string) {
	var titleQuery []string
	var tags []string

	words := strings.Fields(query)
	for _, word := range words {
		if strings.HasPrefix(word, "#") {
			tag := strings.TrimPrefix(word, "#")
			if len(tag) > 0 {
				tags = append(tags, tag)
			}
		} else {
			titleQuery = append(titleQuery, word)
		}
	}

	return strings.Join(titleQuery, " "), tags
}

// DeleteOwnComment permet à un utilisateur de supprimer son propre commentaire
// ou au créateur du post de supprimer un commentaire sur son post
func (h *ForumHandler) DeleteOwnComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Récupérer l'utilisateur connecté
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	// Récupérer l'ID du commentaire
	commentIDStr := r.FormValue("comment_id")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "ID de commentaire invalide", http.StatusBadRequest)
		return
	}

	// Récupérer le commentaire
	comment, err := h.repo.GetCommentByID(commentID)
	if err != nil {
		http.Error(w, "Commentaire non trouvé", http.StatusNotFound)
		return
	}

	// Récupérer le post pour vérifier si l'utilisateur en est le créateur
	post, err := h.repo.GetPostByID(comment.PostID)
	if err != nil {
		http.Error(w, "Post non trouvé", http.StatusNotFound)
		return
	}

	// Vérifier les permissions : propriétaire du commentaire OU créateur du post OU modérateur+
	canDelete := comment.UserID == user.ID || post.UserID == user.ID || user.RoleID >= models.RoleModerator
	if !canDelete {
		http.Error(w, "Permission refusée", http.StatusForbidden)
		return
	}

	// Supprimer le commentaire
	err = h.repo.DeleteComment(commentID)
	if err != nil {
		http.Error(w, "Erreur lors de la suppression", http.StatusInternalServerError)
		return
	}

	// Retourner une réponse de succès
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Commentaire supprimé avec succès"))
}

// DeleteOwnPost permet à un utilisateur de supprimer son propre post
func (h *ForumHandler) DeleteOwnPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Récupérer l'utilisateur connecté
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	// Récupérer l'ID du post
	postIDStr := r.FormValue("post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "ID de post invalide", http.StatusBadRequest)
		return
	}

	// Récupérer le post
	post, err := h.repo.GetPostByID(postID)
	if err != nil {
		http.Error(w, "Post non trouvé", http.StatusNotFound)
		return
	}

	// Vérifier les permissions : propriétaire du post OU modérateur+
	canDelete := post.UserID == user.ID || user.RoleID >= models.RoleModerator
	if !canDelete {
		http.Error(w, "Permission refusée", http.StatusForbidden)
		return
	}

	// Supprimer le post (et ses commentaires en cascade)
	err = h.repo.DeletePost(postID)
	if err != nil {
		http.Error(w, "Erreur lors de la suppression", http.StatusInternalServerError)
		return
	}

	// Rediriger vers la catégorie
	http.Redirect(w, r, fmt.Sprintf("/category/%d", post.CategoryID), http.StatusSeeOther)
}

// === GESTION DES STATUTS DES POSTS ===

// ChangePostStatus permet de changer le statut d'un post (open, closed, archived)
func (h *ForumHandler) ChangePostStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Méthode non autorisée"})
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Non autorisé"})
		return
	}

	// Récupérer les paramètres
	postIDStr := r.FormValue("post_id")
	newStatus := r.FormValue("status")
	reason := r.FormValue("reason")

	// Validation
	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID de post invalide"})
		return
	}

	if newStatus != models.PostStatusOpen && newStatus != models.PostStatusClosed && newStatus != models.PostStatusArchived {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Statut invalide"})
		return
	}

	// Récupérer le post pour vérifier les permissions
	post, err := h.repo.GetPost(postID, user)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Post non trouvé"})
		return
	}

	// Vérifier si l'utilisateur peut changer le statut
	if !post.CanChangeStatusBy(user.ID, user.RoleID) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Permission refusée"})
		return
	}

	// Changer le statut
	err = h.repo.ChangePostStatus(postID, newStatus, user.ID, reason)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur lors du changement de statut"})
		return
	}

	// Retourner JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"message":    "Statut du post mis à jour",
		"new_status": newStatus,
	})
}

// GET /uploads/posts/{filename}
func (h *ForumHandler) ServeImage(w http.ResponseWriter, r *http.Request) {
	// Extraire le nom du fichier depuis l'URL
	path := r.URL.Path
	filename := strings.TrimPrefix(path, "/uploads/posts/")

	// Validation basique du nom de fichier
	if filename == "" || strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		http.Error(w, "Nom de fichier invalide", http.StatusBadRequest)
		return
	}

	// Vérifier que l'image existe en base de données
	_, err := h.repo.GetImageByFilename(filename)
	if err != nil {
		http.Error(w, "Image non trouvée", http.StatusNotFound)
		return
	}

	// Servir le fichier
	filePath := "./uploads/posts/" + filename
	http.ServeFile(w, r, filePath)
}
