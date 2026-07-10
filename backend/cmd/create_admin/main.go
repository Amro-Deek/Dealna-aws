package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/secondary/auth"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/config"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/database"
)

func main() {
	email := flag.String("email", "", "Admin email address")
	password := flag.String("password", "", "Admin password")
	name := flag.String("name", "Dealna Admin", "Admin display name")
	universityID := flag.String("university_id", "", "UUID of the university the admin belongs to")
	flag.Parse()

	if *email == "" || *password == "" || *universityID == "" {
		log.Println("Usage: go run main.go -email admin@dealna.com -password yourpass -university_id <uuid> [-name 'Admin Name']")
		os.Exit(1)
	}

	cfg := config.Load()

	ctx := context.Background()

	// 1. Connect to Database
	db, err := database.Connect(
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 2. Setup Keycloak Identity Provider
	keycloakIdentity := auth.NewKeycloakIdentityProvider(
		cfg.KeycloakBaseURL,
		cfg.KeycloakRealm,
		cfg.KeycloakClientID,
		cfg.KeycloakAdminClientID,
		cfg.KeycloakAdminClientSecret,
		&http.Client{},
	)

	// 3. Register User in Keycloak
	log.Printf("Registering user %s in Keycloak...", *email)
	// Since we only have a generic "name", we split it or just pass it as firstName
	keycloakSub, err := keycloakIdentity.RegisterUser(ctx, *email, *password, *name, "")
	if err != nil {
		log.Fatalf("Failed to register in Keycloak (is the Keycloak server reachable?): %v", err)
	}
	log.Printf("Successfully registered in Keycloak! Sub UUID: %s", keycloakSub)

	// 4. Assign ADMIN Role in Keycloak
	log.Printf("Assigning ADMIN role in Keycloak...")
	err = keycloakIdentity.AssignRoleToUser(ctx, keycloakSub, "ADMIN")
	if err != nil {
		// Attempting fallback to just print warning, but typically this is fatal
		log.Printf("Warning: Failed to assign ADMIN role in Keycloak: %v", err)
	}

	// 5. Insert into Postgres User table
	log.Printf("Inserting user into Postgres...")
	var userID string
	err = db.QueryRow(ctx, `
		INSERT INTO "User" (email, role, account_status, email_verified, university_id, keycloak_sub)
		VALUES ($1, 'ADMIN', 'ACTIVE', true, $2, $3)
		RETURNING user_id
	`, *email, *universityID, keycloakSub).Scan(&userID)
	if err != nil {
		log.Fatalf("Failed to insert into User table: %v", err)
	}

	// 6. Insert into Postgres admin table
	log.Printf("Inserting into admin table...")
	_, err = db.Exec(ctx, `
		INSERT INTO admin (user_id, admin_name)
		VALUES ($1, $2)
	`, userID, *name)
	if err != nil {
		log.Fatalf("Failed to insert into admin table: %v", err)
	}

	log.Printf("✅ Success! Admin account created successfully. User ID: %s", userID)
}
