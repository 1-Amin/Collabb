package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const UserIDKey contextKey = "userID"

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func jwtSecret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		s = "dev-secret-change-in-prod"
	}
	return []byte(s)
}

func SignToken(userID string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret())
}

func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecret(), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

type Service struct{ db *pgxpool.Pool }

func NewService(db *pgxpool.Pool) *Service { return &Service{db: db} }

type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginOutput struct {
	Token string `json:"token"`
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (*LoginOutput, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	var id uuid.UUID
	err = s.db.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		in.Email, string(hash),
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}
	token, err := SignToken(id.String())
	if err != nil {
		return nil, err
	}
	return &LoginOutput{Token: token, ID: id.String(), Email: in.Email}, nil
}

func (s *Service) Login(ctx context.Context, in RegisterInput) (*LoginOutput, error) {
	var id uuid.UUID
	var hash string
	err := s.db.QueryRow(ctx,
		`SELECT id, password_hash FROM users WHERE email = $1`, in.Email,
	).Scan(&id, &hash)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	token, err := SignToken(id.String())
	if err != nil {
		return nil, err
	}
	return &LoginOutput{Token: token, ID: id.String(), Email: in.Email}, nil
}
