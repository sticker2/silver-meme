package auth

import (
	"DistributedCalc/internal/storage"
	"DistributedCalc/pkg/errors"
	"DistributedCalc/pkg/logger"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db   *storage.SQLiteDB
	logr *logger.Logger
}

func NewAuthService(db *storage.SQLiteDB, logr *logger.Logger) *AuthService {
	return &AuthService{db: db, logr: logr}
}

type UserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserClaims struct {
	UserID int64 `json:"user_id"`
	jwt.StandardClaims
}

const (
	UserIDKey = "user_id"
	secretKey = "your-secret-key"
)

func (s *AuthService) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logr.Error("Failed to decode register request: %v", err)
		errors.HandleHTTPError(w, errors.NewBadRequestError("invalid request body"))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logr.Error("Failed to hash password: %v", err)
		errors.HandleHTTPError(w, errors.NewInternalError("failed to register user"))
		return
	}

	userID, err := s.db.CreateUser(req.Login, string(hashedPassword))
	if err != nil {
		s.logr.Error("Failed to create user: %v", err)
		if err.Error() == "user already exists" {
			errors.HandleHTTPError(w, errors.NewBadRequestError("user already exists"))
		} else {
			errors.HandleHTTPError(w, errors.NewInternalError("failed to register user"))
		}
		return
	}

	s.logr.Info("User registered with ID %d", userID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "user registered"})
}

func (s *AuthService) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logr.Error("Failed to decode login request: %v", err)
		errors.HandleHTTPError(w, errors.NewBadRequestError("invalid request body"))
		return
	}

	user, err := s.db.GetUser(req.Login)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		s.logr.Error("Invalid login credentials for user %s: %v", req.Login, err)
		errors.HandleHTTPError(w, errors.NewBadRequestError("invalid credentials"))
		return
	}

	claims := UserClaims{
		UserID: user.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		s.logr.Error("Failed to generate token: %v", err)
		errors.HandleHTTPError(w, errors.NewInternalError("failed to generate token"))
		return
	}

	s.logr.Info("User %d logged in", user.ID)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func (s *AuthService) JWTMiddleware(next http.Handler, authService *AuthService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" || !strings.HasPrefix(tokenStr, "Bearer ") {
			s.logr.Error("Missing or invalid Authorization header")
			errors.HandleHTTPError(w, errors.NewBadRequestError("missing or invalid token"))
			return
		}
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

		claims := &UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
		if err != nil || !token.Valid {
			s.logr.Error("Invalid token: %v", err)
			errors.HandleHTTPError(w, errors.NewBadRequestError("invalid token"))
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
