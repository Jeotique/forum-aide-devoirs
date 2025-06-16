package utils

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashe un mot de passe avec bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash vérifie si un mot de passe correspond au hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GetIntParam récupère un paramètre entier de l'URL
func GetIntParam(r *http.Request, param string) (int, error) {
	value := r.URL.Query().Get(param)
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

// GetStringParam récupère un paramètre string de l'URL
func GetStringParam(r *http.Request, param string) string {
	return r.URL.Query().Get(param)
}

// ExtractIDFromPath extrait l'ID d'un chemin comme "/post/123"
func ExtractIDFromPath(path, prefix string) (int, error) {
	idStr := strings.TrimPrefix(path, prefix)
	return strconv.Atoi(idStr)
}

// SetHTTPOnlyCookie définit un cookie HttpOnly
func SetHTTPOnlyCookie(w http.ResponseWriter, name, value string, maxAge time.Duration) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  time.Now().Add(maxAge),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// DeleteCookie supprime un cookie
func DeleteCookie(w http.ResponseWriter, name string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, cookie)
}

// RenderTemplate rend un template avec gestion d'erreur
func RenderTemplate(w http.ResponseWriter, tmpl *template.Template, name string, data interface{}) {
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "Erreur de rendu du template", http.StatusInternalServerError)
	}
}

// RenderSimplePage rend une page HTML simple en cas d'erreur de template
func RenderSimplePage(w http.ResponseWriter, title, content string) {
	html := `<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + title + ` - Forum d'aide aux devoirs</title>
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" rel="stylesheet">
    <link href="/static/style.css" rel="stylesheet">
</head>
<body>
    <div class="container">
        <div class="simple-page">
            <h1><i class="fas fa-graduation-cap"></i> Forum d'aide aux devoirs</h1>
            <h2>` + title + `</h2>
            ` + content + `
            <div class="nav-links">
                <a href="/" class="btn btn-primary"><i class="fas fa-home"></i> Accueil</a>
                <a href="/login" class="btn btn-secondary"><i class="fas fa-sign-in-alt"></i> Connexion</a>
            </div>
        </div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// SanitizeInput nettoie une entrée utilisateur basique
func SanitizeInput(input string) string {
	// Supprime les espaces en début/fin
	input = strings.TrimSpace(input)

	// Remplace les retours à la ligne multiples
	input = strings.ReplaceAll(input, "\r\n", "\n")

	return input
}

// TruncateText tronque un texte à une longueur donnée
func TruncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
}

// FormatTime formate une date pour l'affichage
func FormatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "À l'instant"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "Il y a 1 minute"
		}
		return "Il y a " + strconv.Itoa(minutes) + " minutes"
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "Il y a 1 heure"
		}
		return "Il y a " + strconv.Itoa(hours) + " heures"
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "Hier"
		}
		return "Il y a " + strconv.Itoa(days) + " jours"
	default:
		return t.Format("02/01/2006 15:04")
	}
}

// IsValidEmail vérifie si un email est valide (validation basique)
func IsValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// IsValidUsername vérifie si un nom d'utilisateur est valide
func IsValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 50 {
		return false
	}

	// Caractères autorisés : lettres, chiffres, underscore, tiret
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return false
		}
	}

	return true
}

// Contains vérifie si une slice contient un élément
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveDuplicates supprime les doublons d'une slice de strings
func RemoveDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// ParseTags parse une chaîne de tags séparés par des virgules
func ParseTags(tagsStr string) []string {
	if tagsStr == "" {
		return []string{}
	}

	tags := strings.Split(tagsStr, ",")
	var cleanTags []string

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" && len(tag) <= 50 {
			cleanTags = append(cleanTags, tag)
		}
	}

	return RemoveDuplicates(cleanTags)
}

// GetClientIP récupère l'IP du client
func GetClientIP(r *http.Request) string {
	// Vérifier les headers de proxy
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	return r.RemoteAddr
}

// FormatContent convertit les retours à la ligne en balises HTML <br>
func FormatContent(content string) template.HTML {
	// Échapper le HTML pour éviter les injections XSS
	content = template.HTMLEscapeString(content)

	// Convertir les retours à la ligne en <br>
	content = strings.ReplaceAll(content, "\n", "<br>")
	content = strings.ReplaceAll(content, "\r", "")

	return template.HTML(content)
}

// ErrorPageData représente les données pour une page d'erreur
type ErrorPageData struct {
	Code    string
	Title   string
	Message string
}

// RenderErrorPage rend une page d'erreur personnalisée
func RenderErrorPage(w http.ResponseWriter, tmpl *template.Template, code int, title, message string) {
	w.WriteHeader(code)

	data := ErrorPageData{
		Code:    strconv.Itoa(code),
		Title:   title,
		Message: message,
	}

	if tmpl != nil {
		if err := tmpl.ExecuteTemplate(w, "error.html", data); err != nil {
			// Fallback en cas d'erreur de template
			RenderSimpleErrorPage(w, code, title, message)
		}
	} else {
		RenderSimpleErrorPage(w, code, title, message)
	}
}

// RenderSimpleErrorPage rend une page d'erreur simple en cas de problème de template
func RenderSimpleErrorPage(w http.ResponseWriter, code int, title, message string) {
	html := `<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + title + ` - Forum d'aide aux devoirs</title>
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" rel="stylesheet">
    <style>
        body { font-family: 'Segoe UI', sans-serif; margin: 0; padding: 20px; background: #f9fafb; }
        .error-container { max-width: 600px; margin: 50px auto; text-align: center; background: white; padding: 40px; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
        .error-code { font-size: 4rem; font-weight: bold; color: #2563eb; margin-bottom: 20px; }
        .error-title { font-size: 1.5rem; font-weight: 600; color: #1f2937; margin-bottom: 15px; }
        .error-message { color: #6b7280; margin-bottom: 30px; line-height: 1.6; }
        .btn { display: inline-block; padding: 12px 24px; background: #2563eb; color: white; text-decoration: none; border-radius: 6px; margin: 0 10px; }
        .btn:hover { background: #1d4ed8; }
        .btn-secondary { background: #6b7280; }
        .btn-secondary:hover { background: #4b5563; }
    </style>
</head>
<body>
    <div class="error-container">
        <div class="error-code">` + strconv.Itoa(code) + `</div>
        <h1 class="error-title">` + title + `</h1>
        <p class="error-message">` + message + `</p>
        <div>
            <a href="/" class="btn"><i class="fas fa-home"></i> Retour à l'accueil</a>
            <a href="javascript:history.back()" class="btn btn-secondary"><i class="fas fa-arrow-left"></i> Page précédente</a>
        </div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// RenderJSONError rend une erreur au format JSON
func RenderJSONError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(`{"error": "` + message + `", "status": "error"}`))
}

// IntPtr retourne un pointeur vers un int
func IntPtr(i int) *int {
	return &i
}
