package auth

import (
	"context"
	"errors"
	"log"
	"task_auth/users"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	service *users.Service
}

func NewAuthService(service *users.Service) *AuthService {
	return &AuthService{
		service: service,
	}
}

func (a *AuthService) LoginUser(ctx context.Context, Code string) (*users.User, error) {
	if Code == "" {
		return nil, errors.New("Code required")
	}

	user, err := a.service.Repo.FindUserCodeDB(ctx, Code)
	if err != nil {
		return nil, errors.New("Database error")
	}
	if user == nil {
		return nil, errors.New("Code doesn't exist")
	}

	accesstoken, err := CreateAccessToken(ctx, user)
	if err != nil {
		log.Printf("Error creating access token: %v", err)
		return nil, errors.New("Failed to create access token")
	}

	refreshtoken, err := CreateRefreshToken(ctx, user)
	if err != nil {
		log.Printf("Error creating refresh token: %v", err)
		return nil, errors.New("Failed to create refresh token")
	}

	user.AccessToken = accesstoken
	user.RefreshToken = refreshtoken

	err = a.service.Repo.SaveUserDB(ctx, user)
	if err != nil {
		log.Printf("Error saving user: %v", err)
		return nil, errors.New("Failed to save user")
	}

	return user, nil
}

func CreateAccessToken(ctx context.Context, user *users.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(60 * time.Minute).Unix(),
		"type":    "access",
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("sigma"))
	if err != nil {
		return "", errors.New("Error with create access token")
	}
	return token, nil
}

func CreateRefreshToken(ctx context.Context, user *users.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"type":    "refresh",
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("sigma"))
	if err != nil {
		return "", errors.New("Error with create refresh token")
	}
	return token, nil
}

func (a *AuthService) AccessTokenValidate(ctx context.Context, refreshtoken string, id int) error {
	user, err := a.service.Repo.FindUserDB(ctx, id)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("User not found")
	}

	if refreshtoken != user.RefreshToken {
		return errors.New("Invalid refresh token")
	}

	accesstoken, err := CreateAccessToken(ctx, user)
	if err != nil {
		return err
	}

	user.AccessToken = accesstoken
	err = a.service.Repo.UpdateUserDB(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (a *AuthService) CreateTeacher(ctx context.Context, user *users.User) (*users.User, error) {
	if user == nil {
		return nil, errors.New("user is nil")
	}

	user.Role = "Teacher"

	created, err := a.service.CreateUser(ctx, user)
	if err != nil {
		log.Printf("❌ Ошибка CreateUser: %v", err)
		return nil, err
	}

	err = a.service.Repo.SaveUserDB(ctx, created)
	if err != nil {
		log.Printf("❌ Ошибка SaveUserDB: %v", err)
		return nil, errors.New("Failed to save user")
	}

	log.Printf("✅ Учитель создан: %s %s (ID=%d)", created.First_name, created.Last_name, created.ID)
	return created, nil
}
func (a *AuthService) GetAllUsers(ctx context.Context) ([]*users.User, error) {
	users, err := a.service.Repo.GetAllUsersDB(ctx)
	if err != nil {
		log.Printf("❌ Ошибка получения пользователей: %v", err)
		return nil, err
	}
	log.Printf("📋 Получено %d пользователей через AuthService", len(users))
	return users, nil
}
