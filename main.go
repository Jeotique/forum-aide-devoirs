package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"

	"aide-devoir-forum/config"
	"aide-devoir-forum/database"
	"aide-devoir-forum/handlers"
	"aide-devoir-forum/middleware"
	"aide-devoir-forum/utils"
)

func main() {
	// Charger la configuration
	cfg := config.Load()

	// Connexion à la base de données
	db, err := sql.Open("mysql", cfg.GetDSN())
	if err != nil {
		log.Fatal("Erreur de connexion à la DB:", err)
	}
	defer db.Close()

	// Tester la connexion
	if err := db.Ping(); err != nil {
		log.Fatal("Impossible de ping la DB:", err)
	}

	fmt.Println("✅ Connecté à la base de données MySQL")

	// Créer le repository
	repo := database.NewRepository(db)

	// Charger les templates
	var templates *template.Template
	templatePath := filepath.Join("templates", "*.html")

	// Créer les fonctions personnalisées pour les templates
	funcMap := template.FuncMap{
		"formatContent": utils.FormatContent,
		"mul":           func(a, b int) int { return a * b },
		"add":           func(a, b int) int { return a + b },
		"dict": func(values ...interface{}) map[string]interface{} {
			dict := make(map[string]interface{})
			for i := 0; i < len(values); i += 2 {
				if i+1 < len(values) {
					dict[values[i].(string)] = values[i+1]
				}
			}
			return dict
		},
		"formatAction": func(action string) string {
			switch action {
			case "ban":
				return "Bannissement d'utilisateur"
			case "unban":
				return "Débannissement d'utilisateur"
			case "promote":
				return "Changement de rôle"
			case "delete_post":
				return "Suppression de post"
			case "delete_comment":
				return "Suppression de commentaire"
			default:
				return "Action " + action
			}
		},
	}

	templates, err = template.New("").Funcs(funcMap).ParseGlob(templatePath)
	if err != nil {
		log.Printf("⚠️  Erreur chargement templates: %v (mode fallback activé)", err)
		templates = nil
	} else {
		fmt.Println("✅ Templates chargés")
	}

	// Créer les handlers
	authHandler := handlers.NewAuthHandler(repo, cfg, templates)
	forumHandler := handlers.NewForumHandler(repo, cfg, templates)
	adminHandler := handlers.NewAdminHandler(repo, cfg, templates)
	profileHandler := handlers.NewProfileHandler(repo, cfg, templates)

	// Créer le serveur HTTP
	mux := http.NewServeMux()

	// Servir les fichiers statiques
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Servir les images uploadées
	mux.HandleFunc("/uploads/posts/", forumHandler.ServeImage)
	mux.HandleFunc("/uploads/avatars/", profileHandler.ServeAvatar)

	// Routes publiques avec middleware optionnel
	mux.HandleFunc("/", middleware.OptionalAuthWithRepo(cfg, repo)(http.HandlerFunc(forumHandler.Home)).ServeHTTP)
	mux.HandleFunc("/search", middleware.OptionalAuthWithRepo(cfg, repo)(http.HandlerFunc(forumHandler.Search)).ServeHTTP)
	mux.HandleFunc("/search-suggestions", middleware.OptionalAuthWithRepo(cfg, repo)(http.HandlerFunc(forumHandler.SearchSuggestions)).ServeHTTP)

	// Routes d'authentification (gèrent GET et POST)
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			authHandler.LoginPage(w, r)
		} else {
			authHandler.Login(w, r)
		}
	})
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			authHandler.RegisterPage(w, r)
		} else {
			authHandler.Register(w, r)
		}
	})
	mux.HandleFunc("/logout", authHandler.Logout)

	// Routes du forum avec middleware optionnel
	mux.HandleFunc("/category/", middleware.OptionalAuthWithRepo(cfg, repo)(http.HandlerFunc(forumHandler.Category)).ServeHTTP)
	mux.HandleFunc("/post/", middleware.OptionalAuthWithRepo(cfg, repo)(http.HandlerFunc(forumHandler.Post)).ServeHTTP)

	// Routes des profils avec middleware optionnel
	mux.HandleFunc("/profile/", middleware.OptionalAuthWithRepo(cfg, repo)(http.HandlerFunc(profileHandler.Profile)).ServeHTTP)

	// Routes nécessitant une authentification
	mux.HandleFunc("/create-post", middleware.RequireAuthWithRepo(cfg, repo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			forumHandler.CreatePostPage(w, r)
		} else {
			forumHandler.CreatePost(w, r)
		}
	})).ServeHTTP)
	mux.HandleFunc("/comment", middleware.RequireAuthWithRepo(cfg, repo)(http.HandlerFunc(forumHandler.CreateComment)).ServeHTTP)
	mux.HandleFunc("/vote", middleware.RequireAuthWithRepo(cfg, repo)(http.HandlerFunc(forumHandler.Vote)).ServeHTTP)
	mux.HandleFunc("/change-post-status", middleware.RequireAuthWithRepo(cfg, repo)(http.HandlerFunc(forumHandler.ChangePostStatus)).ServeHTTP)
	mux.HandleFunc("/mark-solution", middleware.RequireAuthWithRepo(cfg, repo)(http.HandlerFunc(adminHandler.MarkSolution)).ServeHTTP)
	mux.HandleFunc("/delete-own-comment", middleware.RequireAuthWithRepo(cfg, repo)(http.HandlerFunc(forumHandler.DeleteOwnComment)).ServeHTTP)
	mux.HandleFunc("/delete-own-post", middleware.RequireAuthWithRepo(cfg, repo)(http.HandlerFunc(forumHandler.DeleteOwnPost)).ServeHTTP)

	// Routes de gestion du profil (authentifiées)
	mux.HandleFunc("/settings", middleware.RequireAuthWithRepo(cfg, repo)(http.HandlerFunc(profileHandler.Settings)).ServeHTTP)
	mux.HandleFunc("/profile/update", middleware.RequireAuthWithRepo(cfg, repo)(http.HandlerFunc(profileHandler.UpdateProfile)).ServeHTTP)
	mux.HandleFunc("/profile/avatar", middleware.RequireAuthWithRepo(cfg, repo)(http.HandlerFunc(profileHandler.UpdateAvatar)).ServeHTTP)

	// Routes d'administration
	mux.HandleFunc("/admin", middleware.RequireAdminWithRepo(cfg, repo)(http.HandlerFunc(adminHandler.Dashboard)).ServeHTTP)
	mux.HandleFunc("/admin/ban", middleware.RequireModeratorWithRepo(cfg, repo)(http.HandlerFunc(adminHandler.BanUser)).ServeHTTP)
	mux.HandleFunc("/admin/unban", middleware.RequireModeratorWithRepo(cfg, repo)(http.HandlerFunc(adminHandler.UnbanUser)).ServeHTTP)
	mux.HandleFunc("/admin/promote", middleware.RequireAdminWithRepo(cfg, repo)(http.HandlerFunc(adminHandler.PromoteUser)).ServeHTTP)
	mux.HandleFunc("/admin/delete-post", middleware.RequireModeratorWithRepo(cfg, repo)(http.HandlerFunc(adminHandler.DeletePost)).ServeHTTP)
	mux.HandleFunc("/admin/delete-comment", middleware.RequireModeratorWithRepo(cfg, repo)(http.HandlerFunc(adminHandler.DeleteComment)).ServeHTTP)

	// Routes de gestion des catégories
	mux.HandleFunc("/admin/categories", middleware.RequireAdminWithRepo(cfg, repo)(http.HandlerFunc(adminHandler.CreateCategory)).ServeHTTP)
	mux.HandleFunc("/admin/categories/", middleware.RequireAdminWithRepo(cfg, repo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			adminHandler.UpdateCategory(w, r)
		} else if r.Method == "DELETE" {
			adminHandler.DeleteCategory(w, r)
		} else {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		}
	})).ServeHTTP)

	// Appliquer les middlewares globaux
	handler := middleware.Logging()(mux)
	handler = middleware.CORS()(handler)
	handler = middleware.RateLimit()(handler)

	// Configuration du serveur
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Démarrer le serveur
	fmt.Printf("🚀 Serveur démarré sur http://localhost:%s\n", cfg.Server.Port)
	fmt.Println("📚 Forum d'aide aux devoirs prêt !")
	fmt.Println("\n🔗 Liens utiles :")
	fmt.Printf("   Accueil: http://localhost:%s\n", cfg.Server.Port)
	fmt.Printf("   Admin:   http://localhost:%s/admin\n", cfg.Server.Port)
	fmt.Println("\n👤 Comptes de test :")
	fmt.Println("   admin / admin123 (Administrateur)")
	fmt.Println("   prof_martin / prof123 (Professeur)")
	fmt.Println("   eleve_sarah / eleve123 (Élève)")
	fmt.Println()

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Erreur serveur:", err)
	}
}
