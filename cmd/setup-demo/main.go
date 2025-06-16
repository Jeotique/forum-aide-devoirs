package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func main() {
	// Connexion à la base de données
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/forum?charset=utf8mb4&parseTime=true")
	if err != nil {
		log.Fatal("Erreur de connexion à la DB:", err)
	}
	defer db.Close()

	// Tester la connexion
	if err := db.Ping(); err != nil {
		log.Fatal("Impossible de ping la DB:", err)
	}

	fmt.Println("✅ Connecté à la base de données")

	// Supprimer les comptes existants pour éviter les doublons
	fmt.Println("🧹 Nettoyage des anciens comptes de démo...")
	_, err = db.Exec("DELETE FROM users WHERE username IN ('admin', 'prof_martin', 'eleve_sarah')")
	if err != nil {
		log.Printf("Erreur lors du nettoyage: %v", err)
	}

	// Créer les comptes de démonstration
	accounts := []struct {
		username string
		email    string
		password string
		roleID   int
	}{
		{"admin", "admin@forum.local", "admin123", 4},
		{"prof_martin", "prof.martin@forum.local", "prof123", 2},
		{"eleve_sarah", "eleve.sarah@forum.local", "eleve123", 1},
	}

	fmt.Println("🔐 Création des comptes de démonstration...")

	for _, account := range accounts {
		// Hasher le mot de passe
		hashedPassword, err := hashPassword(account.password)
		if err != nil {
			log.Printf("Erreur lors du hashage du mot de passe pour %s: %v", account.username, err)
			continue
		}

		// Insérer l'utilisateur
		_, err = db.Exec(`
			INSERT INTO users (username, email, password, role_id) 
			VALUES (?, ?, ?, ?)
		`, account.username, account.email, hashedPassword, account.roleID)

		if err != nil {
			log.Printf("Erreur lors de la création de %s: %v", account.username, err)
		} else {
			fmt.Printf("✅ Compte créé: %s (mot de passe: %s)\n", account.username, account.password)
		}
	}

	fmt.Println("\n🎉 Comptes de démonstration configurés avec succès !")
	fmt.Println("\n👤 Connexions disponibles :")
	fmt.Println("   admin / admin123 (Administrateur)")
	fmt.Println("   prof_martin / prof123 (Professeur)")
	fmt.Println("   eleve_sarah / eleve123 (Élève)")
}
