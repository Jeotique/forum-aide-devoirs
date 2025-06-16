package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Security SecurityConfig
	Uploads  UploadsConfig
}

type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	Charset  string
}

type JWTConfig struct {
	SecretKey      []byte
	ExpirationTime time.Duration
}

type SecurityConfig struct {
	BCryptCost int
	RateLimit  int
}

type UploadsConfig struct {
	MaxFileSize int64
	PostsDir    string
	AvatarsDir  string
}

func Load() *Config {
	// Charger le fichier .env
	if err := godotenv.Load(); err != nil {
		log.Printf("Aucun fichier .env trouvé, utilisation des variables d'environnement système")
	}

	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Host:         getEnv("SERVER_HOST", "localhost"),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Database: getEnv("DB_NAME", "forum"),
			Charset:  "utf8mb4",
		},
		JWT: JWTConfig{
			SecretKey:      []byte(getEnv("JWT_SECRET", "votre-cle-secrete-jwt-aide-devoir-2024")),
			ExpirationTime: 24 * time.Hour,
		},
		Security: SecurityConfig{
			BCryptCost: getEnvAsInt("BCRYPT_COST", 12),
			RateLimit:  getEnvAsInt("RATE_LIMIT", 100),
		},
		Uploads: UploadsConfig{
			MaxFileSize: getEnvAsInt64("MAX_FILE_SIZE", 10485760), // 10MB par défaut
			PostsDir:    getEnv("UPLOADS_POSTS_DIR", "uploads/posts"),
			AvatarsDir:  getEnv("UPLOADS_AVATARS_DIR", "uploads/avatars"),
		},
	}
}

func (c *Config) GetDSN() string {
	return c.Database.User + ":" + c.Database.Password +
		"@tcp(" + c.Database.Host + ":" + c.Database.Port + ")/" +
		c.Database.Database + "?charset=" + c.Database.Charset + "&parseTime=true"
}

// getEnv récupère une variable d'environnement avec une valeur par défaut
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt récupère une variable d'environnement en tant qu'entier
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsInt64 récupère une variable d'environnement en tant qu'entier 64 bits
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
