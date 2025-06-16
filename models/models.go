package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User représente un utilisateur du forum
type User struct {
	ID                int        `json:"id" db:"id"`
	Username          string     `json:"username" db:"username"`
	Email             string     `json:"email" db:"email"`
	Password          string     `json:"-" db:"password"`
	RoleID            int        `json:"role_id" db:"role_id"`
	RoleName          string     `json:"role_name" db:"role_name"`
	IsBanned          bool       `json:"is_banned" db:"is_banned"`
	BanReason         string     `json:"ban_reason" db:"ban_reason"`
	BannedUntil       time.Time  `json:"banned_until" db:"banned_until"`
	Avatar            string     `json:"avatar" db:"avatar"` // Ancien champ
	Bio               string     `json:"bio" db:"bio"`       // Ancien champ
	AvatarFilename    *string    `json:"avatar_filename" db:"avatar_filename"`
	LastLogin         *time.Time `json:"last_login" db:"last_login"`
	ProfileVisibility string     `json:"profile_visibility" db:"profile_visibility"`
	DateInscription   *time.Time `json:"date_inscription" db:"date_inscription"`
	Location          *string    `json:"location" db:"location"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	AvatarURL         string     `json:"avatar_url"` // URL calculée côté serveur
	Stats             *UserStats `json:"stats,omitempty"`
}

// UserStats représente les statistiques d'un utilisateur
type UserStats struct {
	UserID                int       `json:"user_id" db:"user_id"`
	PostsCount            int       `json:"posts_count" db:"posts_count"`
	CommentsCount         int       `json:"comments_count" db:"comments_count"`
	SolutionsGiven        int       `json:"solutions_given" db:"solutions_given"`
	SolutionsReceived     int       `json:"solutions_received" db:"solutions_received"`
	LikesReceivedPosts    int       `json:"likes_received_posts" db:"likes_received_posts"`
	LikesReceivedComments int       `json:"likes_received_comments" db:"likes_received_comments"`
	TotalViewsPosts       int       `json:"total_views_posts" db:"total_views_posts"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

// Category représente une catégorie de matière
type Category struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Color       string    `json:"color" db:"color"`
	Icon        string    `json:"icon" db:"icon"`
	PostCount   int       `json:"post_count" db:"post_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Tag représente un tag pour organiser les posts
type Tag struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Color     string    `json:"color" db:"color"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Post représente un post du forum
type Post struct {
	ID            int       `json:"id" db:"id"`
	Title         string    `json:"title" db:"title"`
	Content       string    `json:"content" db:"content"`
	UserID        int       `json:"user_id" db:"user_id"`
	Username      string    `json:"username" db:"username"`
	UserRole      string    `json:"user_role" db:"user_role"`
	UserBanned    bool      `json:"user_banned" db:"user_banned"`
	UserAvatarURL string    `json:"user_avatar_url"` // URL calculée côté serveur
	CategoryID    int       `json:"category_id" db:"category_id"`
	CategoryName  string    `json:"category_name" db:"category_name"`
	Status        string    `json:"status" db:"status"`
	IsSolved      bool      `json:"is_solved" db:"is_solved"`
	IsPinned      bool      `json:"is_pinned" db:"is_pinned"`
	IsLocked      bool      `json:"is_locked" db:"is_locked"`
	ViewsCount    int       `json:"views_count" db:"views_count"`
	LikesCount    int       `json:"likes_count" db:"likes_count"`
	DislikesCount int       `json:"dislikes_count" db:"dislikes_count"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	Tags          []Tag     `json:"tags"`
	Images        []Image   `json:"images"`
	UserVote      string    `json:"user_vote"`
}

// Comment représente un commentaire sur un post
type Comment struct {
	ID            int       `json:"id" db:"id"`
	PostID        int       `json:"post_id" db:"post_id"`
	Content       string    `json:"content" db:"content"`
	UserID        int       `json:"user_id" db:"user_id"`
	Username      string    `json:"username" db:"username"`
	UserRole      string    `json:"user_role" db:"user_role"`
	UserBanned    bool      `json:"user_banned" db:"user_banned"`
	UserAvatarURL string    `json:"user_avatar_url"` // URL calculée côté serveur
	ParentID      int       `json:"parent_id" db:"parent_id"`
	IsSolution    bool      `json:"is_solution" db:"is_solution"`
	LikesCount    int       `json:"likes_count" db:"likes_count"`
	DislikesCount int       `json:"dislikes_count" db:"dislikes_count"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	UserVote      string    `json:"user_vote"`
	Replies       []Comment `json:"replies"`
	Images        []Image   `json:"images"`
}

// Vote représente un vote (like/dislike) sur un post ou commentaire
type Vote struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	TargetID  int       `json:"target_id" db:"target_id"`
	VoteType  string    `json:"vote_type" db:"vote_type"` // 'like' ou 'dislike'
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Role représente un rôle d'utilisateur
type Role struct {
	ID          int    `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Permissions string `json:"permissions" db:"permissions"` // JSON
}

// ModerationLog représente un log d'action de modération
type ModerationLog struct {
	ID          int       `json:"id" db:"id"`
	ModeratorID int       `json:"moderator_id" db:"moderator_id"`
	ActionType  string    `json:"action_type" db:"action_type"`
	TargetType  string    `json:"target_type" db:"target_type"`
	TargetID    int       `json:"target_id" db:"target_id"`
	Reason      string    `json:"reason" db:"reason"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Report représente un signalement d'utilisateur
type Report struct {
	ID           int       `json:"id" db:"id"`
	ReporterID   int       `json:"reporter_id" db:"reporter_id"`
	ReportedType string    `json:"reported_type" db:"reported_type"` // 'post', 'comment', 'user'
	ReportedID   int       `json:"reported_id" db:"reported_id"`
	Reason       string    `json:"reason" db:"reason"`
	Description  string    `json:"description" db:"description"`
	Status       string    `json:"status" db:"status"` // 'pending', 'reviewed', 'resolved', 'dismissed'
	ModeratorID  int       `json:"moderator_id" db:"moderator_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	ResolvedAt   time.Time `json:"resolved_at" db:"resolved_at"`
}

// JWT Claims pour l'authentification
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	RoleID   int    `json:"role_id"`
	jwt.RegisteredClaims
}

// Constantes pour les rôles
const (
	RoleUser          = 1
	RoleProfessor     = 2
	RoleModerator     = 3
	RoleAdministrator = 4
)

// Constantes pour les types de votes
const (
	VoteLike    = "like"
	VoteDislike = "dislike"
)

// Constantes pour les statuts de signalement
const (
	ReportStatusPending   = "pending"
	ReportStatusReviewed  = "reviewed"
	ReportStatusResolved  = "resolved"
	ReportStatusDismissed = "dismissed"
)

// Constantes pour les statuts de posts
const (
	PostStatusOpen     = "open"
	PostStatusClosed   = "closed"
	PostStatusArchived = "archived"
)

// Méthodes utilitaires pour User
func (u *User) IsAdmin() bool {
	return u.RoleID >= RoleAdministrator
}

func (u *User) IsModerator() bool {
	return u.RoleID >= RoleModerator
}

func (u *User) IsProfessor() bool {
	return u.RoleID >= RoleProfessor
}

func (u *User) CanModerate() bool {
	return u.RoleID >= RoleModerator && !u.IsBanned
}

func (u *User) CanPromoteUsers() bool {
	return u.RoleID >= RoleAdministrator && !u.IsBanned
}

// Méthodes utilitaires pour Post
func (p *Post) CanBeEditedBy(userID int) bool {
	return p.UserID == userID || !p.IsLocked
}

func (p *Post) GetStatusBadges() []string {
	var badges []string
	if p.IsSolved {
		badges = append(badges, "solved")
	}
	if p.IsPinned {
		badges = append(badges, "pinned")
	}
	if p.IsLocked {
		badges = append(badges, "locked")
	}
	if p.Status == PostStatusClosed {
		badges = append(badges, "closed")
	}
	if p.Status == PostStatusArchived {
		badges = append(badges, "archived")
	}
	return badges
}

// Vérifie si un post peut être vu par un utilisateur
func (p *Post) CanBeViewedBy(userID int, userRoleID int) bool {
	// Les posts archivés ne peuvent être vus que par le propriétaire ou les admins/modérateurs
	if p.Status == PostStatusArchived {
		return p.UserID == userID || userRoleID >= RoleModerator
	}
	// Les posts ouverts et fermés sont visibles par tous
	return true
}

// Vérifie si un post peut recevoir de nouveaux commentaires
func (p *Post) CanReceiveComments() bool {
	return p.Status == PostStatusOpen && !p.IsLocked
}

// Vérifie si un utilisateur peut changer le statut d'un post
func (p *Post) CanChangeStatusBy(userID int, userRoleID int) bool {
	// Le propriétaire peut fermer/rouvrir son propre post
	if p.UserID == userID {
		return true
	}
	// Les modérateurs et admins peuvent changer tous les statuts
	return userRoleID >= RoleModerator
}

// Méthodes utilitaires pour Comment
func (c *Comment) CanBeMarkedAsSolution(userID int, postAuthorID int) bool {
	return postAuthorID == userID && !c.IsSolution
}

// Structures pour les requêtes
type CreatePostRequest struct {
	Title      string `json:"title" form:"title" validate:"required,min=5,max=255"`
	Content    string `json:"content" form:"content" validate:"required,min=20"`
	CategoryID int    `json:"category_id" form:"category_id" validate:"required,min=1"`
	Tags       string `json:"tags" form:"tags"`
}

type CreateCommentRequest struct {
	PostID   int    `json:"post_id" form:"post_id" validate:"required,min=1"`
	Content  string `json:"content" form:"content" validate:"required,min=5"`
	ParentID int    `json:"parent_id" form:"parent_id"`
}

type VoteRequest struct {
	Type     string `json:"type" form:"type" validate:"required,oneof=like dislike"`
	Target   string `json:"target" form:"target" validate:"required,oneof=post comment"`
	TargetID int    `json:"target_id" form:"target_id" validate:"required,min=1"`
}

type LoginRequest struct {
	Username string `json:"username" form:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" form:"password" validate:"required,min=6"`
}

type RegisterRequest struct {
	Username string `json:"username" form:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" form:"email" validate:"required,email"`
	Password string `json:"password" form:"password" validate:"required,min=6"`
}

// Structures pour les réponses
type HomePageData struct {
	Categories  []Category `json:"categories"`
	RecentPosts []Post     `json:"recent_posts"`
	User        *User      `json:"user"`
	Title       string     `json:"title"`
}

type CategoryPageData struct {
	Category Category `json:"category"`
	Posts    []Post   `json:"posts"`
	User     *User    `json:"user"`
	Title    string   `json:"title"`
}

type PostPageData struct {
	Post           Post         `json:"post"`
	Comments       []Comment    `json:"comments"`
	User           *User        `json:"user"`
	Title          string       `json:"title"`
	CurrentSort    string       `json:"current_sort"`
	AvailableSorts []SortOption `json:"available_sorts"`
}

type SortOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type AdminPageData struct {
	Users      []User          `json:"users"`
	Categories []Category      `json:"categories"`
	Logs       []ModerationLog `json:"logs"`
	Stats      AdminStats      `json:"stats"`
	User       *User           `json:"user"`
	Title      string          `json:"title"`
}

type AdminStats struct {
	BannedUsers    int `json:"banned_users"`
	Administrators int `json:"administrators"`
	Professors     int `json:"professors"`
}

// Image représente une image uploadée
type Image struct {
	ID           int       `json:"id" db:"id"`
	Filename     string    `json:"filename" db:"filename"`
	OriginalName string    `json:"original_name" db:"original_name"`
	ContentType  string    `json:"content_type" db:"content_type"`
	SizeBytes    int       `json:"size_bytes" db:"size_bytes"`
	Width        int       `json:"width" db:"width"`
	Height       int       `json:"height" db:"height"`
	PostID       *int      `json:"post_id" db:"post_id"`
	CommentID    *int      `json:"comment_id" db:"comment_id"`
	UserID       int       `json:"user_id" db:"user_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	URL          string    `json:"url"` // URL calculée côté serveur
}

// UserActivity représente une activité récente d'un utilisateur
type UserActivity struct {
	Type      string    `json:"type"` // "post", "comment", "solution"
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	PostID    int       `json:"post_id,omitempty"`
	PostTitle string    `json:"post_title,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ProfilePageData représente les données pour la page de profil
type ProfilePageData struct {
	ProfileUser    User           `json:"profile_user"`
	IsOwnProfile   bool           `json:"is_own_profile"`
	CanViewProfile bool           `json:"can_view_profile"`
	RecentActivity []UserActivity `json:"recent_activity"`
	User           *User          `json:"user"` // Utilisateur connecté
	Title          string         `json:"title"`
}

// SettingsPageData représente les données pour la page de paramètres
type SettingsPageData struct {
	User  *User  `json:"user"`
	Title string `json:"title"`
}

// UpdateProfileRequest représente une demande de mise à jour de profil
type UpdateProfileRequest struct {
	Bio               string `json:"bio" form:"bio" validate:"max=500"`
	Location          string `json:"location" form:"location" validate:"max=100"`
	ProfileVisibility string `json:"profile_visibility" form:"profile_visibility" validate:"oneof=public private"`
}
