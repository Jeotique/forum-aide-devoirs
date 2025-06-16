package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"aide-devoir-forum/config"
	"aide-devoir-forum/models"
	"aide-devoir-forum/utils"
)

type contextKey string

const UserContextKey = contextKey("user")

// RequireAuth middleware qui vérifie la présence d'un token JWT valide
func RequireAuth(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromRequest(r, cfg)
			if user == nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuthWithRepo middleware qui vérifie la présence d'un token JWT valide avec accès au repository
func RequireAuthWithRepo(cfg *config.Config, repo interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromRequestWithRepo(r, cfg, repo)
			if user == nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole middleware qui vérifie que l'utilisateur a un rôle minimum
func RequireRole(cfg *config.Config, minRole int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromRequest(r, cfg)

			// Vérifier si c'est une requête AJAX
			isAjax := strings.Contains(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") ||
				strings.Contains(r.Header.Get("Accept"), "application/json") ||
				r.Header.Get("X-Requested-With") == "XMLHttpRequest"

			if user == nil {
				if isAjax {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(map[string]string{
						"status": "error",
						"error":  "Accès refusé - utilisateur non trouvé (pas de token ou token invalide)",
					})
				} else {
					http.Error(w, "Accès refusé - utilisateur non trouvé", http.StatusForbidden)
				}
				return
			}

			if user.RoleID < minRole {
				if isAjax {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(map[string]string{
						"status": "error",
						"error":  "Accès refusé - rôle insuffisant (RoleID: " + fmt.Sprintf("%d", user.RoleID) + ", requis: " + fmt.Sprintf("%d", minRole) + ")",
					})
				} else {
					http.Error(w, "Accès refusé - rôle insuffisant", http.StatusForbidden)
				}
				return
			}

			if user.IsBanned {
				if isAjax {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(map[string]string{
						"status": "error",
						"error":  "Votre compte est banni: " + user.BanReason,
					})
				} else {
					http.Error(w, "Votre compte est banni: "+user.BanReason, http.StatusForbidden)
				}
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth middleware qui ajoute l'utilisateur au contexte s'il est connecté
func OptionalAuth(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromRequest(r, cfg)

			var ctx context.Context
			if user != nil {
				ctx = context.WithValue(r.Context(), UserContextKey, user)
			} else {
				ctx = r.Context()
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthWithRepo middleware qui ajoute l'utilisateur au contexte s'il est connecté avec accès au repository
func OptionalAuthWithRepo(cfg *config.Config, repo interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromRequestWithRepo(r, cfg, repo)

			var ctx context.Context
			if user != nil {
				ctx = context.WithValue(r.Context(), UserContextKey, user)
			} else {
				ctx = r.Context()
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromRequest extrait l'utilisateur du token JWT dans la requête
func GetUserFromRequest(r *http.Request, cfg *config.Config) *models.User {
	cookie, err := r.Cookie("token")
	if err != nil {
		// Debug: pas de cookie token
		return nil
	}

	claims, err := utils.ParseJWTToken(cookie.Value, cfg.JWT.SecretKey)
	if err != nil {
		// Debug: token invalide
		return nil
	}

	// Créer un utilisateur basique à partir des claims
	// Dans un vrai système, vous devriez récupérer l'utilisateur complet de la DB
	user := &models.User{
		ID:       claims.UserID,
		Username: claims.Username,
		RoleID:   claims.RoleID,
		IsBanned: false, // Par défaut, on considère que l'utilisateur n'est pas banni
	}

	// Debug: vérifier que l'utilisateur a bien été créé
	if user.ID == 0 || user.RoleID == 0 {
		return nil
	}

	return user
}

// GetUserFromRequestWithRepo extrait l'utilisateur du token JWT et le récupère depuis la DB
func GetUserFromRequestWithRepo(r *http.Request, cfg *config.Config, repo interface{}) *models.User {
	cookie, err := r.Cookie("token")
	if err != nil {
		return nil
	}

	claims, err := utils.ParseJWTToken(cookie.Value, cfg.JWT.SecretKey)
	if err != nil {
		return nil
	}

	// Interface pour le repository
	type UserRepository interface {
		GetUserByIDComplete(id int) (*models.User, error)
	}

	userRepo, ok := repo.(UserRepository)
	if !ok {
		// Fallback vers l'ancienne méthode si le repository n'est pas valide
		return GetUserFromRequest(r, cfg)
	}

	// Récupérer l'utilisateur complet depuis la base de données
	user, err := userRepo.GetUserByIDComplete(claims.UserID)
	if err != nil {
		return nil
	}

	return user
}

// GetUserFromContext récupère l'utilisateur du contexte de la requête
func GetUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

// RequireModerator middleware pour les actions de modération
func RequireModerator(cfg *config.Config) func(http.Handler) http.Handler {
	return RequireRole(cfg, models.RoleModerator)
}

// RequireAdmin middleware pour les actions d'administration
func RequireAdmin(cfg *config.Config) func(http.Handler) http.Handler {
	return RequireRole(cfg, models.RoleAdministrator)
}

// RequireRoleWithRepo middleware qui vérifie que l'utilisateur a un rôle minimum avec accès au repository
func RequireRoleWithRepo(cfg *config.Config, repo interface{}, minRole int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromRequestWithRepo(r, cfg, repo)

			// Vérifier si c'est une requête AJAX
			isAjax := strings.Contains(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") ||
				strings.Contains(r.Header.Get("Accept"), "application/json") ||
				r.Header.Get("X-Requested-With") == "XMLHttpRequest"

			if user == nil {
				if isAjax {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(map[string]string{
						"status": "error",
						"error":  "Accès refusé - utilisateur non trouvé (pas de token ou token invalide)",
					})
				} else {
					http.Error(w, "Accès refusé - utilisateur non trouvé", http.StatusForbidden)
				}
				return
			}

			if user.RoleID < minRole {
				if isAjax {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(map[string]string{
						"status": "error",
						"error":  "Accès refusé - rôle insuffisant (RoleID: " + fmt.Sprintf("%d", user.RoleID) + ", requis: " + fmt.Sprintf("%d", minRole) + ")",
					})
				} else {
					http.Error(w, "Accès refusé - rôle insuffisant", http.StatusForbidden)
				}
				return
			}

			if user.IsBanned {
				if isAjax {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(map[string]string{
						"status": "error",
						"error":  "Votre compte est banni: " + user.BanReason,
					})
				} else {
					http.Error(w, "Votre compte est banni: "+user.BanReason, http.StatusForbidden)
				}
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireModeratorWithRepo middleware pour les actions de modération avec repository
func RequireModeratorWithRepo(cfg *config.Config, repo interface{}) func(http.Handler) http.Handler {
	return RequireRoleWithRepo(cfg, repo, models.RoleModerator)
}

// RequireAdminWithRepo middleware pour les actions d'administration avec repository
func RequireAdminWithRepo(cfg *config.Config, repo interface{}) func(http.Handler) http.Handler {
	return RequireRoleWithRepo(cfg, repo, models.RoleAdministrator)
}

// CORS middleware pour les requêtes AJAX
func CORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Logging middleware pour logger les requêtes
func Logging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// Logger basique - vous pouvez utiliser une bibliothèque comme logrus
		// log.Printf("%s %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		return next
	}
}

// RateLimit middleware basique pour limiter les requêtes
// Dans un vrai système, utilisez une bibliothèque comme golang.org/x/time/rate
func RateLimit() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// Implémentation basique - à améliorer en production
		return next
	}
}
