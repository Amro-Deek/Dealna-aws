package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/Amro-Deek/Dealna-aws/internal/database"
	"github.com/Amro-Deek/Dealna-aws/internal/database/generated"
	"github.com/Amro-Deek/Dealna-aws/internal/middleware"
	"github.com/Amro-Deek/Dealna-aws/internal/models"
	"github.com/Amro-Deek/Dealna-aws/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
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

	pool := database.GetPool()
	if pool == nil {
		utils.WriteJSON(w, http.StatusInternalServerError, false, "Database not initialized", nil, nil)
		return
	}

	queries := generated.New(pool)

	user, err := queries.GetUserForLogin(context.Background(), req.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			utils.WriteJSON(w, http.StatusUnauthorized, false, "Invalid credentials", nil, nil)
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, false, "Database error", nil, err)
		return
	}

	if !user.PasswordHash.Valid {
		utils.WriteJSON(w, http.StatusUnauthorized, false, "Invalid credentials", nil, nil)
		return
	}

	if bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash.String),
		[]byte(req.Password),
	) != nil {
		utils.WriteJSON(w, http.StatusUnauthorized, false, "Invalid credentials", nil, nil)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.UserID.String(),
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

func GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.ContextUserID).(string)
	if !ok || userID == "" {
		utils.WriteJSON(w, http.StatusUnauthorized, false, "Unauthorized", nil, "Missing user_id")
		return
	}

	role, ok := r.Context().Value(middleware.ContextRole).(string)
	if !ok || role == "" {
		utils.WriteJSON(w, http.StatusUnauthorized, false, "Unauthorized", nil, "Missing role")
		return
	}

	utils.WriteJSON(
		w,
		http.StatusOK,
		true,
		"Profile fetched",
		models.MeResponse{
			UserID: userID,
			Role:   role,
		},
		nil,
	)
}

