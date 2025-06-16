package database

import (
	"database/sql"
	"fmt"
	"strings"

	"aide-devoir-forum/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// === USERS ===

func (r *Repository) GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(`
		SELECT u.id, u.username, u.email, u.password, u.role_id, r.name, 
		       u.is_banned, COALESCE(u.ban_reason, '') as ban_reason, u.created_at
		FROM users u 
		JOIN roles r ON u.role_id = r.id 
		WHERE u.username = ?`, username).
		Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.RoleID, &user.RoleName,
			&user.IsBanned, &user.BanReason, &user.CreatedAt)
	return user, err
}

func (r *Repository) GetUserByID(id int) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(`
		SELECT u.id, u.username, u.email, u.role_id, r.name, 
		       u.is_banned, COALESCE(u.ban_reason, '') as ban_reason, u.created_at
		FROM users u 
		JOIN roles r ON u.role_id = r.id 
		WHERE u.id = ?`, id).
		Scan(&user.ID, &user.Username, &user.Email, &user.RoleID, &user.RoleName,
			&user.IsBanned, &user.BanReason, &user.CreatedAt)
	return user, err
}

// GetUserByIDComplete récupère un utilisateur avec plus de détails (version compatible)
func (r *Repository) GetUserByIDComplete(id int) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(`
		SELECT u.id, u.username, u.email, u.role_id, r.name, 
		       u.is_banned, COALESCE(u.ban_reason, '') as ban_reason, 
		       COALESCE(u.avatar, '') as avatar, COALESCE(u.bio, '') as bio,
		       u.avatar_filename, u.last_login, COALESCE(u.profile_visibility, 'public') as profile_visibility, 
		       u.date_inscription, u.location, u.created_at
		FROM users u 
		JOIN roles r ON u.role_id = r.id 
		WHERE u.id = ?`, id).
		Scan(&user.ID, &user.Username, &user.Email, &user.RoleID, &user.RoleName,
			&user.IsBanned, &user.BanReason, &user.Avatar, &user.Bio,
			&user.AvatarFilename, &user.LastLogin, &user.ProfileVisibility,
			&user.DateInscription, &user.Location, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	// Construire l'URL de l'avatar
	if user.AvatarFilename != nil && *user.AvatarFilename != "" {
		user.AvatarURL = "/uploads/avatars/" + *user.AvatarFilename
	}

	return user, err
}

func (r *Repository) CreateUser(username, email, hashedPassword string) error {
	_, err := r.db.Exec("INSERT INTO users (username, email, password, role_id) VALUES (?, ?, ?, 1)",
		username, email, hashedPassword)
	return err
}

func (r *Repository) GetAllUsers() ([]models.User, error) {
	rows, err := r.db.Query(`
		SELECT u.id, u.username, u.email, u.role_id, r.name, 
		       u.is_banned, COALESCE(u.ban_reason, '') as ban_reason, u.created_at
		FROM users u 
		JOIN roles r ON u.role_id = r.id 
		ORDER BY u.created_at DESC
		LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.RoleID, &user.RoleName,
			&user.IsBanned, &user.BanReason, &user.CreatedAt)
		if err != nil {
			continue
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *Repository) BanUser(userID int, reason string) error {
	_, err := r.db.Exec("UPDATE users SET is_banned = TRUE, ban_reason = ? WHERE id = ?", reason, userID)
	return err
}

func (r *Repository) PromoteUser(userID, newRoleID int) error {
	_, err := r.db.Exec("UPDATE users SET role_id = ? WHERE id = ?", newRoleID, userID)
	return err
}

// UnbanUser débannit un utilisateur
func (r *Repository) UnbanUser(userID int) error {
	query := `UPDATE users SET is_banned = FALSE, ban_reason = '', banned_until = NULL WHERE id = ?`
	_, err := r.db.Exec(query, userID)
	return err
}

// === CATEGORIES ===

func (r *Repository) GetCategories() ([]models.Category, error) {
	rows, err := r.db.Query(`
		SELECT c.id, c.name, c.description, c.color, c.icon, COUNT(p.id) as post_count
		FROM categories c
		LEFT JOIN posts p ON c.id = p.category_id
		GROUP BY c.id, c.name, c.description, c.color, c.icon
		ORDER BY c.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		err := rows.Scan(&category.ID, &category.Name, &category.Description,
			&category.Color, &category.Icon, &category.PostCount)
		if err != nil {
			continue
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (r *Repository) GetCategory(id int) (*models.Category, error) {
	category := &models.Category{}
	err := r.db.QueryRow(`
		SELECT c.id, c.name, c.description, c.color, c.icon, COUNT(p.id) as post_count
		FROM categories c
		LEFT JOIN posts p ON c.id = p.category_id
		WHERE c.id = ?
		GROUP BY c.id, c.name, c.description, c.color, c.icon
	`, id).Scan(&category.ID, &category.Name, &category.Description,
		&category.Color, &category.Icon, &category.PostCount)
	return category, err
}

// === POSTS ===

func (r *Repository) GetRecentPosts(limit int) ([]models.Post, error) {
	rows, err := r.db.Query(`
		SELECT p.id, p.title, LEFT(p.content, 200) as content, p.user_id, u.username, r.name as role_name, u.is_banned, u.avatar_filename,
		       p.category_id, c.name as category_name, p.status, p.is_solved, p.is_pinned, p.is_locked,
		       p.views_count, p.likes_count, p.dislikes_count, p.created_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN roles r ON u.role_id = r.id
		JOIN categories c ON p.category_id = c.id
		WHERE p.status != 'archived'
		ORDER BY p.created_at DESC 
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var avatarFilename sql.NullString
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.Username, &post.UserRole, &post.UserBanned, &avatarFilename,
			&post.CategoryID, &post.CategoryName, &post.Status, &post.IsSolved, &post.IsPinned, &post.IsLocked,
			&post.ViewsCount, &post.LikesCount, &post.DislikesCount, &post.CreatedAt)
		if err != nil {
			continue
		}

		// Calculer l'URL de l'avatar
		if avatarFilename.Valid {
			post.UserAvatarURL = "/uploads/avatars/" + avatarFilename.String
		}

		posts = append(posts, post)
	}
	return posts, nil
}

func (r *Repository) GetPostsByCategory(categoryID int) ([]models.Post, error) {
	rows, err := r.db.Query(`
		SELECT p.id, p.title, LEFT(p.content, 200) as content, p.user_id, u.username, r.name as role_name, u.is_banned, u.avatar_filename,
		       p.category_id, c.name as category_name, p.status, p.is_solved, p.is_pinned, p.is_locked,
		       p.views_count, p.likes_count, p.dislikes_count, p.created_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN roles r ON u.role_id = r.id
		JOIN categories c ON p.category_id = c.id
		WHERE p.category_id = ? AND p.status != 'archived'
		ORDER BY p.is_pinned DESC, p.created_at DESC
	`, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var avatarFilename sql.NullString
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.Username, &post.UserRole, &post.UserBanned, &avatarFilename,
			&post.CategoryID, &post.CategoryName, &post.Status, &post.IsSolved, &post.IsPinned, &post.IsLocked,
			&post.ViewsCount, &post.LikesCount, &post.DislikesCount, &post.CreatedAt)
		if err != nil {
			continue
		}

		// Calculer l'URL de l'avatar
		if avatarFilename.Valid {
			post.UserAvatarURL = "/uploads/avatars/" + avatarFilename.String
		}

		posts = append(posts, post)
	}
	return posts, nil
}

func (r *Repository) GetPost(id int, user *models.User) (*models.Post, error) {
	post := &models.Post{}
	var avatarFilename sql.NullString
	err := r.db.QueryRow(`
		SELECT p.id, p.title, p.content, p.user_id, u.username, r.name as role_name, u.is_banned, u.avatar_filename,
		       p.category_id, c.name as category_name, p.status, p.is_solved, p.is_pinned, p.is_locked,
		       p.views_count, p.likes_count, p.dislikes_count, p.created_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN roles r ON u.role_id = r.id
		JOIN categories c ON p.category_id = c.id
		WHERE p.id = ?
	`, id).Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.Username, &post.UserRole, &post.UserBanned, &avatarFilename,
		&post.CategoryID, &post.CategoryName, &post.Status, &post.IsSolved, &post.IsPinned, &post.IsLocked,
		&post.ViewsCount, &post.LikesCount, &post.DislikesCount, &post.CreatedAt)

	if err != nil {
		return nil, err
	}

	// Calculer l'URL de l'avatar
	if avatarFilename.Valid {
		post.UserAvatarURL = "/uploads/avatars/" + avatarFilename.String
	}

	// Charger le vote de l'utilisateur si connecté
	if user != nil {
		var voteType sql.NullString
		r.db.QueryRow("SELECT vote_type FROM post_votes WHERE post_id = ? AND user_id = ?",
			post.ID, user.ID).Scan(&voteType)
		if voteType.Valid {
			post.UserVote = voteType.String
		}
	}

	// Charger les tags
	post.Tags, _ = r.GetPostTags(post.ID)

	// Charger les images
	post.Images, _ = r.GetPostImages(post.ID)

	return post, nil
}

func (r *Repository) CreatePost(title, content string, userID, categoryID int) (int64, error) {
	result, err := r.db.Exec("INSERT INTO posts (title, content, user_id, category_id) VALUES (?, ?, ?, ?)",
		title, content, userID, categoryID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) IncrementPostViews(postID int) error {
	_, err := r.db.Exec("UPDATE posts SET views_count = views_count + 1 WHERE id = ?", postID)
	return err
}

func (r *Repository) MarkPostAsSolved(postID int) error {
	_, err := r.db.Exec("UPDATE posts SET is_solved = TRUE WHERE id = ?", postID)
	return err
}

func (r *Repository) DeletePost(postID int) error {
	_, err := r.db.Exec("DELETE FROM posts WHERE id = ?", postID)
	return err
}

// === COMMENTS ===

func (r *Repository) GetComments(postID int, user *models.User) ([]models.Comment, error) {
	return r.GetCommentsWithSort(postID, user, "newest")
}

func (r *Repository) GetCommentsWithSort(postID int, user *models.User, sortBy string) ([]models.Comment, error) {
	// Définir l'ordre SQL selon le type de tri
	var orderClause string
	switch sortBy {
	case "oldest":
		orderClause = "ORDER BY c.is_solution DESC, c.created_at ASC"
	case "newest":
		orderClause = "ORDER BY c.is_solution DESC, c.created_at DESC"
	case "most_liked":
		orderClause = "ORDER BY c.is_solution DESC, c.likes_count DESC, c.created_at DESC"
	case "solutions_first":
		orderClause = "ORDER BY c.is_solution DESC, c.likes_count DESC, c.created_at ASC"
	default:
		orderClause = "ORDER BY c.is_solution DESC, c.created_at DESC"
	}

	query := fmt.Sprintf(`
		SELECT c.id, c.post_id, c.content, c.user_id, u.username, r.name as role_name, u.is_banned, u.avatar_filename,
		       COALESCE(c.parent_id, 0) as parent_id, c.is_solution, c.likes_count, c.dislikes_count, c.created_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		JOIN roles r ON u.role_id = r.id
		WHERE c.post_id = ?
		%s
	`, orderClause)

	rows, err := r.db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allComments []models.Comment
	commentMap := make(map[int]*models.Comment)

	for rows.Next() {
		var comment models.Comment
		var avatarFilename sql.NullString
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.Content, &comment.UserID,
			&comment.Username, &comment.UserRole, &comment.UserBanned, &avatarFilename, &comment.ParentID, &comment.IsSolution,
			&comment.LikesCount, &comment.DislikesCount, &comment.CreatedAt)
		if err != nil {
			continue
		}

		// Calculer l'URL de l'avatar
		if avatarFilename.Valid {
			comment.UserAvatarURL = "/uploads/avatars/" + avatarFilename.String
		}

		// Charger le vote utilisateur si connecté
		if user != nil {
			var voteType sql.NullString
			r.db.QueryRow("SELECT vote_type FROM comment_votes WHERE comment_id = ? AND user_id = ?",
				comment.ID, user.ID).Scan(&voteType)
			if voteType.Valid {
				comment.UserVote = voteType.String
			}
		}

		// Initialiser le slice des réponses
		comment.Replies = []models.Comment{}

		allComments = append(allComments, comment)
		commentMap[comment.ID] = &allComments[len(allComments)-1]
	}

	// Charger les images pour tous les commentaires
	for i := range allComments {
		comment := &allComments[i]
		images, _ := r.GetCommentImages(comment.ID)
		comment.Images = images

		// Mettre à jour la référence dans la map
		commentMap[comment.ID] = comment
	}

	// Organiser les commentaires en arbre hiérarchique
	// D'abord, attacher les réponses aux parents
	for i := range allComments {
		comment := &allComments[i]
		if comment.ParentID != 0 {
			// Réponse à un commentaire
			if parent, exists := commentMap[comment.ParentID]; exists {
				parent.Replies = append(parent.Replies, *comment)
			}
		}
	}

	// Ensuite, collecter les commentaires racines
	var rootComments []models.Comment
	for i := range allComments {
		comment := &allComments[i]
		if comment.ParentID == 0 {
			// Commentaire racine - prendre la version mise à jour depuis commentMap
			if updatedComment, exists := commentMap[comment.ID]; exists {
				rootComments = append(rootComments, *updatedComment)
			}
		}
	}

	return rootComments, nil
}

func (r *Repository) CreateComment(postID int, content string, userID int, parentID *int) error {
	_, err := r.db.Exec("INSERT INTO comments (post_id, content, user_id, parent_id) VALUES (?, ?, ?, ?)",
		postID, content, userID, parentID)
	return err
}

// CreateCommentWithID crée un commentaire et retourne son ID
func (r *Repository) CreateCommentWithID(postID int, content string, userID int, parentID *int) (int64, error) {
	result, err := r.db.Exec("INSERT INTO comments (post_id, content, user_id, parent_id) VALUES (?, ?, ?, ?)",
		postID, content, userID, parentID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) MarkCommentAsSolution(commentID int) error {
	_, err := r.db.Exec("UPDATE comments SET is_solution = TRUE WHERE id = ?", commentID)
	return err
}

func (r *Repository) DeleteComment(commentID int) error {
	_, err := r.db.Exec("DELETE FROM comments WHERE id = ?", commentID)
	return err
}

// === TAGS ===

func (r *Repository) GetPostTags(postID int) ([]models.Tag, error) {
	rows, err := r.db.Query(`
		SELECT t.id, t.name, t.color
		FROM tags t
		JOIN post_tags pt ON t.id = pt.tag_id
		WHERE pt.post_id = ?
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Color)
		if err != nil {
			continue
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (r *Repository) AddTagToPost(postID int, tagName string) error {
	// Créer le tag s'il n'existe pas
	var tagID int
	err := r.db.QueryRow("SELECT id FROM tags WHERE name = ?", tagName).Scan(&tagID)
	if err != nil {
		result, err := r.db.Exec("INSERT INTO tags (name) VALUES (?)", tagName)
		if err != nil {
			return err
		}
		id, _ := result.LastInsertId()
		tagID = int(id)
	}

	// Lier le tag au post
	_, err = r.db.Exec("INSERT IGNORE INTO post_tags (post_id, tag_id) VALUES (?, ?)", postID, tagID)
	return err
}

// === VOTES ===

func (r *Repository) VotePost(postID, userID int, voteType string) error {
	// Vérifier le vote existant
	var existingVote string
	err := r.db.QueryRow("SELECT vote_type FROM post_votes WHERE post_id = ? AND user_id = ?",
		postID, userID).Scan(&existingVote)

	if err == nil {
		// L'utilisateur a déjà voté
		if existingVote == voteType {
			// Même type de vote → supprimer (toggle off)
			_, err = r.db.Exec("DELETE FROM post_votes WHERE post_id = ? AND user_id = ?",
				postID, userID)
		} else {
			// Type de vote différent → changer le vote
			_, err = r.db.Exec("UPDATE post_votes SET vote_type = ? WHERE post_id = ? AND user_id = ?",
				voteType, postID, userID)
		}
	} else {
		// Pas de vote existant → créer un nouveau vote
		_, err = r.db.Exec("INSERT INTO post_votes (post_id, user_id, vote_type) VALUES (?, ?, ?)",
			postID, userID, voteType)
	}

	if err != nil {
		return err
	}

	// Mettre à jour les compteurs
	return r.UpdatePostVoteCounts(postID)
}

func (r *Repository) VoteComment(commentID, userID int, voteType string) error {
	// Vérifier le vote existant
	var existingVote string
	err := r.db.QueryRow("SELECT vote_type FROM comment_votes WHERE comment_id = ? AND user_id = ?",
		commentID, userID).Scan(&existingVote)

	if err == nil {
		// L'utilisateur a déjà voté
		if existingVote == voteType {
			// Même type de vote → supprimer (toggle off)
			_, err = r.db.Exec("DELETE FROM comment_votes WHERE comment_id = ? AND user_id = ?",
				commentID, userID)
		} else {
			// Type de vote différent → changer le vote
			_, err = r.db.Exec("UPDATE comment_votes SET vote_type = ? WHERE comment_id = ? AND user_id = ?",
				voteType, commentID, userID)
		}
	} else {
		// Pas de vote existant → créer un nouveau vote
		_, err = r.db.Exec("INSERT INTO comment_votes (comment_id, user_id, vote_type) VALUES (?, ?, ?)",
			commentID, userID, voteType)
	}

	if err != nil {
		return err
	}

	// Mettre à jour les compteurs
	return r.UpdateCommentVoteCounts(commentID)
}

func (r *Repository) UpdatePostVoteCounts(postID int) error {
	var likes, dislikes int
	r.db.QueryRow("SELECT COUNT(*) FROM post_votes WHERE post_id = ? AND vote_type = 'like'", postID).Scan(&likes)
	r.db.QueryRow("SELECT COUNT(*) FROM post_votes WHERE post_id = ? AND vote_type = 'dislike'", postID).Scan(&dislikes)
	_, err := r.db.Exec("UPDATE posts SET likes_count = ?, dislikes_count = ? WHERE id = ?", likes, dislikes, postID)
	return err
}

func (r *Repository) UpdateCommentVoteCounts(commentID int) error {
	var likes, dislikes int
	r.db.QueryRow("SELECT COUNT(*) FROM comment_votes WHERE comment_id = ? AND vote_type = 'like'", commentID).Scan(&likes)
	r.db.QueryRow("SELECT COUNT(*) FROM comment_votes WHERE comment_id = ? AND vote_type = 'dislike'", commentID).Scan(&dislikes)
	_, err := r.db.Exec("UPDATE comments SET likes_count = ?, dislikes_count = ? WHERE id = ?", likes, dislikes, commentID)
	return err
}

// === MODERATION ===

func (r *Repository) CreateModerationLog(moderatorID int, actionType, targetType string, targetID int, reason string) error {
	_, err := r.db.Exec(`
		INSERT INTO moderation_logs (moderator_id, action_type, target_type, target_id, reason) 
		VALUES (?, ?, ?, ?, ?)`,
		moderatorID, actionType, targetType, targetID, reason)
	return err
}

// LogModerationAction enregistre une action de modération
func (r *Repository) LogModerationAction(moderatorID int, actionType, targetType string, targetID int, reason string) error {
	query := `INSERT INTO moderation_logs (moderator_id, action_type, target_type, target_id, reason, created_at) 
			  VALUES (?, ?, ?, ?, ?, NOW())`
	_, err := r.db.Exec(query, moderatorID, actionType, targetType, targetID, reason)
	return err
}

// === HELPERS ===

func (r *Repository) GetPostAuthorID(postID int) (int, error) {
	var authorID int
	err := r.db.QueryRow("SELECT user_id FROM posts WHERE id = ?", postID).Scan(&authorID)
	return authorID, err
}

// === RECHERCHE ===

func (r *Repository) SearchPosts(query string, categoryID int, limit int) ([]models.Post, error) {
	return r.SearchPostsAdvanced(query, categoryID, limit)
}

// SearchPostsAdvanced effectue une recherche intelligente par titre et/ou tags
func (r *Repository) SearchPostsAdvanced(query string, categoryID int, limit int) ([]models.Post, error) {
	if query == "" {
		// Si pas de recherche, retourner les posts récents
		if categoryID > 0 {
			return r.GetPostsByCategory(categoryID)
		}
		return r.GetRecentPosts(limit)
	}

	// Analyser la requête pour détecter les tags (#hashtag)
	titleQuery, tags := r.parseSearchQuery(query)

	var whereClause []string
	var args []interface{}
	var joins []string

	baseJoin := `
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN roles r ON u.role_id = r.id
		JOIN categories c ON p.category_id = c.id
	`

	// Recherche par tags
	if len(tags) > 0 {
		joins = append(joins, "JOIN post_tags pt ON p.id = pt.post_id")
		joins = append(joins, "JOIN tags t ON pt.tag_id = t.id")

		tagConditions := make([]string, len(tags))
		for i, tag := range tags {
			tagConditions[i] = "t.name LIKE ?"
			args = append(args, "%"+tag+"%")
		}
		whereClause = append(whereClause, "("+strings.Join(tagConditions, " OR ")+")")
	}

	// Recherche par titre/contenu
	if titleQuery != "" {
		whereClause = append(whereClause, "(p.title LIKE ? OR p.content LIKE ?)")
		searchTerm := "%" + titleQuery + "%"
		args = append(args, searchTerm, searchTerm)
	}

	// Filtrer par catégorie si spécifiée
	if categoryID > 0 {
		whereClause = append(whereClause, "p.category_id = ?")
		args = append(args, categoryID)
	}

	// Exclure les posts archivés des résultats publics
	whereClause = append(whereClause, "p.status != 'archived'")

	// Construire la requête finale
	joinSQL := baseJoin + strings.Join(joins, " ")
	whereSQL := "WHERE " + strings.Join(whereClause, " AND ")

	query = fmt.Sprintf(`
		SELECT DISTINCT p.id, p.title, LEFT(p.content, 200) as content, p.user_id, u.username, r.name as role_name, u.is_banned,
		       p.category_id, c.name as category_name, p.status, p.is_solved, p.is_pinned, p.is_locked,
		       p.views_count, p.likes_count, p.dislikes_count, p.created_at
		%s
		%s
		ORDER BY p.is_pinned DESC, p.created_at DESC 
		LIMIT ?
	`, joinSQL, whereSQL)

	args = append(args, limit)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.Username, &post.UserRole, &post.UserBanned,
			&post.CategoryID, &post.CategoryName, &post.Status, &post.IsSolved, &post.IsPinned, &post.IsLocked,
			&post.ViewsCount, &post.LikesCount, &post.DislikesCount, &post.CreatedAt)
		if err != nil {
			continue
		}

		// Charger les tags pour chaque post
		post.Tags, _ = r.GetPostTags(post.ID)

		posts = append(posts, post)
	}
	return posts, nil
}

// parseSearchQuery analyse une requête de recherche et sépare les tags du texte
func (r *Repository) parseSearchQuery(query string) (string, []string) {
	var titleQuery []string
	var tags []string

	// Diviser la requête en mots
	words := strings.Fields(query)

	for _, word := range words {
		// Si le mot commence par #, c'est un tag
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

// GetPopularTags retourne les tags les plus utilisés
func (r *Repository) GetPopularTags(limit int) ([]models.Tag, error) {
	rows, err := r.db.Query(`
		SELECT t.id, t.name, t.color, COUNT(pt.post_id) as usage_count
		FROM tags t
		JOIN post_tags pt ON t.id = pt.tag_id
		JOIN posts p ON pt.post_id = p.id
		WHERE p.status != 'archived'
		GROUP BY t.id, t.name, t.color
		ORDER BY usage_count DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		var usageCount int
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &usageCount)
		if err != nil {
			continue
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// SearchSuggestions retourne des suggestions de recherche
func (r *Repository) SearchSuggestions(query string, limit int) ([]string, error) {
	if len(query) < 2 {
		return []string{}, nil
	}

	var suggestions []string

	// Suggestions de titres
	rows, err := r.db.Query(`
		SELECT DISTINCT p.title
		FROM posts p
		WHERE p.title LIKE ? AND p.status != 'archived'
		ORDER BY p.views_count DESC, p.created_at DESC
		LIMIT ?
	`, "%"+query+"%", limit/2)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var title string
			if rows.Scan(&title) == nil {
				suggestions = append(suggestions, title)
			}
		}
	}

	// Suggestions de tags
	tagRows, err := r.db.Query(`
		SELECT DISTINCT t.name
		FROM tags t
		JOIN post_tags pt ON t.id = pt.tag_id
		JOIN posts p ON pt.post_id = p.id
		WHERE t.name LIKE ? AND p.status != 'archived'
		ORDER BY (SELECT COUNT(*) FROM post_tags pt2 WHERE pt2.tag_id = t.id) DESC
		LIMIT ?
	`, "%"+query+"%", limit/2)

	if err == nil {
		defer tagRows.Close()
		for tagRows.Next() {
			var tagName string
			if tagRows.Scan(&tagName) == nil {
				suggestions = append(suggestions, "#"+tagName)
			}
		}
	}

	return suggestions, nil
}

// GetCommentByID récupère un commentaire par son ID
func (r *Repository) GetCommentByID(commentID int) (*models.Comment, error) {
	comment := &models.Comment{}
	query := `
		SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, c.likes_count, c.dislikes_count, c.is_solution,
		       u.username
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = ?`

	err := r.db.QueryRow(query, commentID).Scan(
		&comment.ID, &comment.PostID, &comment.UserID, &comment.Content,
		&comment.CreatedAt, &comment.LikesCount, &comment.DislikesCount, &comment.IsSolution,
		&comment.Username,
	)

	if err != nil {
		return nil, err
	}

	return comment, nil
}

// GetPostByID récupère un post par son ID
func (r *Repository) GetPostByID(postID int) (*models.Post, error) {
	post := &models.Post{}
	query := `
		SELECT p.id, p.title, p.content, p.user_id, p.category_id, p.created_at, 
		       p.likes_count, p.dislikes_count, p.is_solved, p.views_count,
		       u.username, c.name as category_name
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN categories c ON p.category_id = c.id
		WHERE p.id = ?`

	err := r.db.QueryRow(query, postID).Scan(
		&post.ID, &post.Title, &post.Content, &post.UserID, &post.CategoryID,
		&post.CreatedAt, &post.LikesCount, &post.DislikesCount, &post.IsSolved, &post.ViewsCount,
		&post.Username, &post.CategoryName,
	)

	if err != nil {
		return nil, err
	}

	return post, nil
}

// === POST STATUS MANAGEMENT ===

// ChangePostStatus change le statut d'un post
func (r *Repository) ChangePostStatus(postID int, status string, moderatorID int, reason string) error {
	// Mettre à jour le statut
	_, err := r.db.Exec("UPDATE posts SET status = ? WHERE id = ?", status, postID)
	if err != nil {
		return err
	}

	// Déterminer le type d'action pour le log
	var actionType string
	switch status {
	case "closed":
		actionType = "close_post"
	case "open":
		actionType = "reopen_post"
	case "archived":
		actionType = "archive_post"
	}

	// Enregistrer l'action de modération
	if actionType != "" {
		r.LogModerationAction(moderatorID, actionType, "post", postID, reason)
	}

	return nil
}

// GetPostsWithAllStatuses récupère les posts pour les admins/modérateurs (incluant archivés)
func (r *Repository) GetPostsWithAllStatuses(userID int, userRoleID int, categoryID int, limit int) ([]models.Post, error) {
	var whereClause string
	var args []interface{}

	if categoryID > 0 {
		whereClause = "WHERE p.category_id = ?"
		args = append(args, categoryID)

		// Si l'utilisateur n'est pas modérateur, exclure les posts archivés sauf les siens
		if userRoleID < models.RoleModerator {
			whereClause += " AND (p.status != 'archived' OR p.user_id = ?)"
			args = append(args, userID)
		}
	} else {
		// Si l'utilisateur n'est pas modérateur, exclure les posts archivés sauf les siens
		if userRoleID < models.RoleModerator {
			whereClause = "WHERE (p.status != 'archived' OR p.user_id = ?)"
			args = append(args, userID)
		}
	}

	args = append(args, limit)

	query := fmt.Sprintf(`
		SELECT p.id, p.title, LEFT(p.content, 200) as content, p.user_id, u.username, r.name as role_name, u.is_banned,
		       p.category_id, c.name as category_name, p.status, p.is_solved, p.is_pinned, p.is_locked,
		       p.views_count, p.likes_count, p.dislikes_count, p.created_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN roles r ON u.role_id = r.id
		JOIN categories c ON p.category_id = c.id
		%s
		ORDER BY p.is_pinned DESC, p.created_at DESC 
		LIMIT ?
	`, whereClause)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.Username, &post.UserRole, &post.UserBanned,
			&post.CategoryID, &post.CategoryName, &post.Status, &post.IsSolved, &post.IsPinned, &post.IsLocked,
			&post.ViewsCount, &post.LikesCount, &post.DislikesCount, &post.CreatedAt)
		if err != nil {
			continue
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// === IMAGES ===

// CreateImage enregistre une nouvelle image en base de données
func (r *Repository) CreateImage(image *models.Image) error {
	query := `
		INSERT INTO images (filename, original_name, content_type, size_bytes, width, height, post_id, comment_id, user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query,
		image.Filename, image.OriginalName, image.ContentType, image.SizeBytes,
		image.Width, image.Height, image.PostID, image.CommentID, image.UserID)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	image.ID = int(id)
	return nil
}

// GetPostImages récupère toutes les images d'un post
func (r *Repository) GetPostImages(postID int) ([]models.Image, error) {
	rows, err := r.db.Query(`
		SELECT id, filename, original_name, content_type, size_bytes, width, height, 
		       post_id, comment_id, user_id, created_at
		FROM images 
		WHERE post_id = ? 
		ORDER BY created_at ASC
	`, postID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		var img models.Image
		err := rows.Scan(&img.ID, &img.Filename, &img.OriginalName, &img.ContentType,
			&img.SizeBytes, &img.Width, &img.Height, &img.PostID, &img.CommentID,
			&img.UserID, &img.CreatedAt)
		if err != nil {
			continue
		}

		// Construire l'URL de l'image
		img.URL = "/uploads/posts/" + img.Filename
		images = append(images, img)
	}

	return images, nil
}

// GetCommentImages récupère toutes les images d'un commentaire
func (r *Repository) GetCommentImages(commentID int) ([]models.Image, error) {
	rows, err := r.db.Query(`
		SELECT id, filename, original_name, content_type, size_bytes, width, height, 
		       post_id, comment_id, user_id, created_at
		FROM images 
		WHERE comment_id = ? 
		ORDER BY created_at ASC
	`, commentID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		var img models.Image
		err := rows.Scan(&img.ID, &img.Filename, &img.OriginalName, &img.ContentType,
			&img.SizeBytes, &img.Width, &img.Height, &img.PostID, &img.CommentID,
			&img.UserID, &img.CreatedAt)
		if err != nil {
			continue
		}

		// Construire l'URL de l'image
		img.URL = "/uploads/posts/" + img.Filename
		images = append(images, img)
	}

	return images, nil
}

// GetImageByID récupère une image par son ID
func (r *Repository) GetImageByID(id int) (*models.Image, error) {
	img := &models.Image{}
	err := r.db.QueryRow(`
		SELECT id, filename, original_name, content_type, size_bytes, width, height,
		       post_id, comment_id, user_id, created_at
		FROM images WHERE id = ?
	`, id).Scan(&img.ID, &img.Filename, &img.OriginalName, &img.ContentType,
		&img.SizeBytes, &img.Width, &img.Height, &img.PostID, &img.CommentID,
		&img.UserID, &img.CreatedAt)

	if err != nil {
		return nil, err
	}

	img.URL = "/uploads/posts/" + img.Filename
	return img, nil
}

// GetImageByFilename récupère une image par son nom de fichier
func (r *Repository) GetImageByFilename(filename string) (*models.Image, error) {
	img := &models.Image{}
	err := r.db.QueryRow(`
		SELECT id, filename, original_name, content_type, size_bytes, width, height,
		       post_id, comment_id, user_id, created_at
		FROM images WHERE filename = ?
	`, filename).Scan(&img.ID, &img.Filename, &img.OriginalName, &img.ContentType,
		&img.SizeBytes, &img.Width, &img.Height, &img.PostID, &img.CommentID,
		&img.UserID, &img.CreatedAt)

	if err != nil {
		return nil, err
	}

	img.URL = "/uploads/posts/" + img.Filename
	return img, nil
}

// DeleteImage supprime une image de la base de données
func (r *Repository) DeleteImage(id int) error {
	_, err := r.db.Exec("DELETE FROM images WHERE id = ?", id)
	return err
}

// GetUserImages récupère toutes les images d'un utilisateur
func (r *Repository) GetUserImages(userID int, limit int) ([]models.Image, error) {
	rows, err := r.db.Query(`
		SELECT id, filename, original_name, content_type, size_bytes, width, height,
		       post_id, comment_id, user_id, created_at
		FROM images 
		WHERE user_id = ? 
		ORDER BY created_at DESC 
		LIMIT ?
	`, userID, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		var img models.Image
		err := rows.Scan(&img.ID, &img.Filename, &img.OriginalName, &img.ContentType,
			&img.SizeBytes, &img.Width, &img.Height, &img.PostID, &img.CommentID,
			&img.UserID, &img.CreatedAt)
		if err != nil {
			continue
		}

		img.URL = "/uploads/posts/" + img.Filename
		images = append(images, img)
	}

	return images, nil
}

// === PROFILS UTILISATEUR ===

// GetUserProfile récupère le profil complet d'un utilisateur avec ses statistiques
func (r *Repository) GetUserProfile(username string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT u.id, u.username, u.email, u.role_id, r.name as role_name, u.is_banned,
		       u.ban_reason, u.banned_until, u.avatar, u.bio, u.avatar_filename,
		       u.last_login, u.profile_visibility, u.date_inscription, u.location, u.created_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.username = ?
	`

	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.RoleID, &user.RoleName,
		&user.IsBanned, &user.BanReason, &user.BannedUntil, &user.Avatar, &user.Bio,
		&user.AvatarFilename, &user.LastLogin, &user.ProfileVisibility,
		&user.DateInscription, &user.Location, &user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Construire l'URL de l'avatar
	if user.AvatarFilename != nil && *user.AvatarFilename != "" {
		user.AvatarURL = "/uploads/avatars/" + *user.AvatarFilename
	}

	// Charger les statistiques
	stats, err := r.GetUserStats(user.ID)
	if err == nil {
		user.Stats = stats
	}

	return user, nil
}

// GetUserStats récupère les statistiques d'un utilisateur
func (r *Repository) GetUserStats(userID int) (*models.UserStats, error) {
	stats := &models.UserStats{}
	query := `
		SELECT user_id, posts_count, comments_count, solutions_given, solutions_received,
		       likes_received_posts, likes_received_comments, total_views_posts,
		       created_at, updated_at
		FROM user_stats
		WHERE user_id = ?
	`

	err := r.db.QueryRow(query, userID).Scan(
		&stats.UserID, &stats.PostsCount, &stats.CommentsCount, &stats.SolutionsGiven,
		&stats.SolutionsReceived, &stats.LikesReceivedPosts, &stats.LikesReceivedComments,
		&stats.TotalViewsPosts, &stats.CreatedAt, &stats.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// UpdateUserProfile met à jour le profil d'un utilisateur
func (r *Repository) UpdateUserProfile(userID int, bio, location, visibility string) error {
	query := `
		UPDATE users 
		SET bio = ?, location = ?, profile_visibility = ?
		WHERE id = ?
	`

	// Gérer location NULL si vide
	var locationValue interface{}
	if location == "" {
		locationValue = nil
	} else {
		locationValue = location
	}

	fmt.Printf("DEBUG UpdateUserProfile: userID=%d, bio='%s', location='%s', visibility='%s'\n",
		userID, bio, location, visibility)
	fmt.Printf("DEBUG UpdateUserProfile: locationValue=%v\n", locationValue)

	result, err := r.db.Exec(query, bio, locationValue, visibility, userID)
	if err != nil {
		fmt.Printf("DEBUG UpdateUserProfile: Erreur SQL: %v\n", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("DEBUG UpdateUserProfile: Lignes affectées: %d\n", rowsAffected)

	return nil
}

// UpdateUserAvatar met à jour l'avatar d'un utilisateur
func (r *Repository) UpdateUserAvatar(userID int, avatarFilename string) error {
	query := `UPDATE users SET avatar_filename = ? WHERE id = ?`
	_, err := r.db.Exec(query, avatarFilename, userID)
	return err
}

// UpdateLastLogin met à jour la dernière connexion d'un utilisateur
func (r *Repository) UpdateLastLogin(userID int) error {
	query := `UPDATE users SET last_login = NOW() WHERE id = ?`
	_, err := r.db.Exec(query, userID)
	return err
}

// GetUserActivity récupère l'activité récente d'un utilisateur
func (r *Repository) GetUserActivity(userID int, limit int) ([]models.UserActivity, error) {
	var activities []models.UserActivity

	// Récupérer les posts récents
	postQuery := `
		SELECT 'post' as type, p.id, p.title, LEFT(p.content, 150) as content, 
		       0 as post_id, '' as post_title, p.created_at
		FROM posts p
		WHERE p.user_id = ?
		ORDER BY p.created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(postQuery, userID, limit/2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var activity models.UserActivity
		err := rows.Scan(&activity.Type, &activity.ID, &activity.Title, &activity.Content,
			&activity.PostID, &activity.PostTitle, &activity.CreatedAt)
		if err != nil {
			continue
		}
		activities = append(activities, activity)
	}

	// Récupérer les commentaires récents
	commentQuery := `
		SELECT 'comment' as type, c.id, '' as title, LEFT(c.content, 150) as content,
		       c.post_id, p.title as post_title, c.created_at
		FROM comments c
		JOIN posts p ON c.post_id = p.id
		WHERE c.user_id = ?
		ORDER BY c.created_at DESC
		LIMIT ?
	`

	rows2, err := r.db.Query(commentQuery, userID, limit/2)
	if err != nil {
		return activities, nil // Retourner au moins les posts
	}
	defer rows2.Close()

	for rows2.Next() {
		var activity models.UserActivity
		err := rows2.Scan(&activity.Type, &activity.ID, &activity.Title, &activity.Content,
			&activity.PostID, &activity.PostTitle, &activity.CreatedAt)
		if err != nil {
			continue
		}
		activities = append(activities, activity)
	}

	// Trier par date décroissante et limiter
	if len(activities) > limit {
		activities = activities[:limit]
	}

	return activities, nil
}

// GetUserPosts récupère les posts d'un utilisateur
func (r *Repository) GetUserPosts(userID int, limit int) ([]models.Post, error) {
	query := `
		SELECT p.id, p.title, LEFT(p.content, 200) as content, p.user_id, u.username,
		       p.category_id, c.name as category_name, p.status, p.is_solved,
		       p.views_count, p.likes_count, p.dislikes_count, p.created_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN categories c ON p.category_id = c.id
		WHERE p.user_id = ?
		ORDER BY p.created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.UserID, &post.Username,
			&post.CategoryID, &post.CategoryName, &post.Status, &post.IsSolved,
			&post.ViewsCount, &post.LikesCount, &post.DislikesCount, &post.CreatedAt)
		if err != nil {
			continue
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// GetUserComments récupère les commentaires d'un utilisateur
func (r *Repository) GetUserComments(userID int, limit int) ([]models.Comment, error) {
	query := `
		SELECT c.id, c.post_id, c.content, c.user_id, u.username,
		       c.is_solution, c.likes_count, c.dislikes_count, c.created_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.user_id = ?
		ORDER BY c.created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.Content, &comment.UserID,
			&comment.Username, &comment.IsSolution, &comment.LikesCount,
			&comment.DislikesCount, &comment.CreatedAt)
		if err != nil {
			continue
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// UpdateUserStats met à jour les statistiques d'un utilisateur
func (r *Repository) UpdateUserStats(userID int) error {
	// Compter les posts
	var postsCount int
	r.db.QueryRow("SELECT COUNT(*) FROM posts WHERE user_id = ?", userID).Scan(&postsCount)

	// Compter les commentaires
	var commentsCount int
	r.db.QueryRow("SELECT COUNT(*) FROM comments WHERE user_id = ?", userID).Scan(&commentsCount)

	// Compter les solutions données
	var solutionsGiven int
	r.db.QueryRow("SELECT COUNT(*) FROM comments WHERE user_id = ? AND is_solution = TRUE", userID).Scan(&solutionsGiven)

	// Compter les solutions reçues (posts résolus)
	var solutionsReceived int
	r.db.QueryRow("SELECT COUNT(*) FROM posts WHERE user_id = ? AND is_solved = TRUE", userID).Scan(&solutionsReceived)

	// Compter les likes reçus sur les posts
	var likesReceivedPosts int
	r.db.QueryRow(`
		SELECT COALESCE(SUM(p.likes_count), 0)
		FROM posts p
		WHERE p.user_id = ?
	`, userID).Scan(&likesReceivedPosts)

	// Compter les likes reçus sur les commentaires
	var likesReceivedComments int
	r.db.QueryRow(`
		SELECT COALESCE(SUM(c.likes_count), 0)
		FROM comments c
		WHERE c.user_id = ?
	`, userID).Scan(&likesReceivedComments)

	// Compter les vues totales des posts
	var totalViewsPosts int
	r.db.QueryRow(`
		SELECT COALESCE(SUM(p.views_count), 0)
		FROM posts p
		WHERE p.user_id = ?
	`, userID).Scan(&totalViewsPosts)

	// Mettre à jour ou insérer
	query := `
		INSERT INTO user_stats (
			user_id, posts_count, comments_count, solutions_given, solutions_received,
			likes_received_posts, likes_received_comments, total_views_posts
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			posts_count = VALUES(posts_count),
			comments_count = VALUES(comments_count),
			solutions_given = VALUES(solutions_given),
			solutions_received = VALUES(solutions_received),
			likes_received_posts = VALUES(likes_received_posts),
			likes_received_comments = VALUES(likes_received_comments),
			total_views_posts = VALUES(total_views_posts),
			updated_at = NOW()
	`

	_, err := r.db.Exec(query, userID, postsCount, commentsCount, solutionsGiven,
		solutionsReceived, likesReceivedPosts, likesReceivedComments, totalViewsPosts)

	return err
}

// === ADMIN METHODS ===

func (r *Repository) GetModerationLogs(limit int) ([]models.ModerationLog, error) {
	rows, err := r.db.Query(`
		SELECT id, moderator_id, action_type, target_type, target_id, reason, created_at
		FROM moderation_logs
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.ModerationLog
	for rows.Next() {
		var log models.ModerationLog
		err := rows.Scan(&log.ID, &log.ModeratorID, &log.ActionType, &log.TargetType,
			&log.TargetID, &log.Reason, &log.CreatedAt)
		if err != nil {
			continue
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func (r *Repository) GetAdminStats() (models.AdminStats, error) {
	var stats models.AdminStats

	// Compter les utilisateurs bannis
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE is_banned = 1").Scan(&stats.BannedUsers)
	if err != nil {
		stats.BannedUsers = 0
	}

	// Compter les administrateurs
	err = r.db.QueryRow("SELECT COUNT(*) FROM users WHERE role_id = 4").Scan(&stats.Administrators)
	if err != nil {
		stats.Administrators = 0
	}

	// Compter les professeurs
	err = r.db.QueryRow("SELECT COUNT(*) FROM users WHERE role_id = 2").Scan(&stats.Professors)
	if err != nil {
		stats.Professors = 0
	}

	return stats, nil
}

func (r *Repository) CreateCategory(name, description, color, icon string) error {
	_, err := r.db.Exec(`
		INSERT INTO categories (name, description, color, icon, created_at)
		VALUES (?, ?, ?, ?, NOW())
	`, name, description, color, icon)
	return err
}

func (r *Repository) UpdateCategory(id int, name, description, color, icon string) error {
	_, err := r.db.Exec(`
		UPDATE categories 
		SET name = ?, description = ?, color = ?, icon = ?
		WHERE id = ?
	`, name, description, color, icon, id)
	return err
}

func (r *Repository) DeleteCategory(id int) error {
	// Vérifier s'il y a des posts dans cette catégorie
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM posts WHERE category_id = ?", id).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("impossible de supprimer une catégorie contenant des posts")
	}

	_, err = r.db.Exec("DELETE FROM categories WHERE id = ?", id)
	return err
}
