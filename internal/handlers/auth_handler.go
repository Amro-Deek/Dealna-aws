package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/Amro-Deek/Dealna-aws/internal/database"
	"github.com/Amro-Deek/Dealna-aws/internal/model"
	"github.com/Amro-Deek/Dealna-aws/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, false, "Invalid request body", nil, err)
		return
	}

	if req.Email == "" || req.Password == "" {
		utils.WriteJSON(w, http.StatusBadRequest, false, "Email and password required", nil, nil)
		return
	}

	db := database.GetDB()
	if db == nil {
		utils.WriteJSON(w, http.StatusInternalServerError, false, "Database not initialized", nil, nil)
		return
	}

	var user models.User
	var hashedPassword string

	err := db.QueryRow(`
		SELECT user_id, email, password_hash, role
		FROM "User"
		WHERE email = $1
	`, req.Email).Scan(&user.ID, &user.Email, &hashedPassword, &user.Role)

	if err == sql.ErrNoRows {
		utils.WriteJSON(w, http.StatusUnauthorized, false, "Invalid credentials", nil, nil)
		return
	}
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, false, "Database error", nil, err)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)) != nil {
		utils.WriteJSON(w, http.StatusUnauthorized, false, "Invalid credentials", nil, nil)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	secret := os.Getenv("JWT_SECRET")
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, false, "Token generation failed", nil, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, true, "Login successful", map[string]string{
		"token": tokenStr,
	}, nil)
}
