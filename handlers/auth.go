package handlers

import (
	"html/template"
	"net/http"

	"aide-devoir-forum/config"
	"aide-devoir-forum/database"
	"aide-devoir-forum/utils"
)

type AuthHandler struct {
	repo      *database.Repository
	config    *config.Config
	templates *template.Template
}

func NewAuthHandler(repo *database.Repository, cfg *config.Config, tmpl *template.Template) *AuthHandler {
	return &AuthHandler{
		repo:      repo,
		config:    cfg,
		templates: tmpl,
	}
}

// GET /login
func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if h.templates != nil {
		utils.RenderTemplate(w, h.templates, "login.html", nil)
	} else {
		utils.RenderSimplePage(w, "Connexion", `
			<form method="POST" action="/login" class="form-container">
				<div class="form-group">
					<label for="username">Nom d'utilisateur :</label>
					<input type="text" id="username" name="username" required>
				</div>
				<div class="form-group">
					<label for="password">Mot de passe :</label>
					<input type="password" id="password" name="password" required>
				</div>
				<button type="submit" class="btn btn-primary">Se connecter</button>
			</form>
			<p><a href="/register">Créer un compte</a></p>
			<div class="demo-accounts" style="margin-top: 20px; padding: 15px; background: #f5f5f5; border-radius: 5px;">
				<h3>Comptes de démonstration :</h3>
				<p><strong>Admin :</strong> admin / admin123</p>
				<p><strong>Prof :</strong> prof_martin / prof123</p>
				<p><strong>Élève :</strong> eleve_sarah / eleve123</p>
			</div>
		`)
	}
}

// POST /login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/login?error=method", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Redirect(w, r, "/login?error=missing", http.StatusSeeOther)
		return
	}

	// Récupérer l'utilisateur
	user, err := h.repo.GetUserByUsername(username)
	if err != nil {
		http.Redirect(w, r, "/login?error=invalid", http.StatusSeeOther)
		return
	}

	// Vérifier le mot de passe
	if !utils.CheckPasswordHash(password, user.Password) {
		http.Redirect(w, r, "/login?error=invalid", http.StatusSeeOther)
		return
	}

	// Vérifier si l'utilisateur est banni
	if user.IsBanned {
		http.Redirect(w, r, "/login?error=banned&reason="+user.BanReason, http.StatusSeeOther)
		return
	}

	// Générer le token JWT
	token, err := utils.GenerateJWTToken(user.ID, user.Username, user.RoleID,
		h.config.JWT.SecretKey, h.config.JWT.ExpirationTime)
	if err != nil {
		http.Redirect(w, r, "/login?error=token", http.StatusSeeOther)
		return
	}

	// Définir le cookie
	utils.SetHTTPOnlyCookie(w, "token", token, h.config.JWT.ExpirationTime)

	// Rediriger vers l'accueil avec message de succès
	http.Redirect(w, r, "/?success=login", http.StatusSeeOther)
}

// GET /register
func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if h.templates != nil {
		utils.RenderTemplate(w, h.templates, "register.html", nil)
	} else {
		utils.RenderSimplePage(w, "Inscription", `
			<form method="POST" action="/register" class="form-container">
				<div class="form-group">
					<label for="username">Nom d'utilisateur :</label>
					<input type="text" id="username" name="username" required minlength="3" maxlength="50">
					<small>3-50 caractères, lettres, chiffres, _ et - autorisés</small>
				</div>
				<div class="form-group">
					<label for="email">Email :</label>
					<input type="email" id="email" name="email" required>
				</div>
				<div class="form-group">
					<label for="password">Mot de passe :</label>
					<input type="password" id="password" name="password" required minlength="6">
					<small>Minimum 6 caractères</small>
				</div>
				<button type="submit" class="btn btn-primary">S'inscrire</button>
			</form>
			<p><a href="/login">Déjà un compte ? Se connecter</a></p>
		`)
	}
}

// POST /register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/register?error=method", http.StatusSeeOther)
		return
	}

	username := utils.SanitizeInput(r.FormValue("username"))
	email := utils.SanitizeInput(r.FormValue("email"))
	password := r.FormValue("password")

	// Validation
	if !utils.IsValidUsername(username) {
		http.Redirect(w, r, "/register?error=username", http.StatusSeeOther)
		return
	}

	if !utils.IsValidEmail(email) {
		http.Redirect(w, r, "/register?error=email", http.StatusSeeOther)
		return
	}

	if len(password) < 6 {
		http.Redirect(w, r, "/register?error=password", http.StatusSeeOther)
		return
	}

	// Vérifier si l'utilisateur existe déjà
	existingUser, _ := h.repo.GetUserByUsername(username)
	if existingUser != nil {
		http.Redirect(w, r, "/register?error=exists", http.StatusSeeOther)
		return
	}

	// Hasher le mot de passe
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		http.Redirect(w, r, "/register?error=hash", http.StatusSeeOther)
		return
	}

	// Créer l'utilisateur
	err = h.repo.CreateUser(username, email, hashedPassword)
	if err != nil {
		http.Redirect(w, r, "/register?error=create", http.StatusSeeOther)
		return
	}

	// Rediriger vers la page de connexion avec succès
	http.Redirect(w, r, "/login?success=register", http.StatusSeeOther)
}

// POST /logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/login?error=method", http.StatusSeeOther)
		return
	}

	// Supprimer le cookie
	utils.DeleteCookie(w, "token")

	// Rediriger vers la page de connexion avec message
	http.Redirect(w, r, "/login?success=logout", http.StatusSeeOther)
}
 