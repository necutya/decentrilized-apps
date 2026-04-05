package service

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/necutya/decentrilized_apps/lab1/internal/model"
	"github.com/necutya/decentrilized_apps/lab1/internal/repo"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users *repo.UserRepo
}

func NewAuthService(users *repo.UserRepo) *AuthService {
	return &AuthService{users: users}
}

func (s *AuthService) Register(username, password, email string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	u := &model.User{Username: username, Email: email, PasswordHash: string(hash)}
	if err := s.users.Create(u); err != nil {
		return "", errors.New("username or email already taken")
	}
	return s.issueToken(u.ID, username)
}

func (s *AuthService) Login(username, password string) (string, error) {
	u, err := s.users.FindByUsername(username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}
	return s.issueToken(u.ID, username)
}

func (s *AuthService) ValidateToken(tokenStr string) (uint, string, error) {
	secret := jwtSecret()
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return 0, "", errors.New("invalid token")
	}
	claims := token.Claims.(*jwt.MapClaims)
	sub, err := claims.GetSubject()
	if err != nil {
		return 0, "", err
	}
	// sub format: "<id>"
	var id uint
	if _, err := fmt.Sscan(sub, &id); err != nil {
		return 0, "", err
	}
	username, _ := (*claims)["username"].(string)
	return id, username, nil
}

func (s *AuthService) issueToken(id uint, username string) (string, error) {
	claims := jwt.MapClaims{
		"sub":      fmt.Sprint(id),
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret())
}

func jwtSecret() []byte {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return []byte(s)
	}
	return []byte("dev-secret-change-me")
}
