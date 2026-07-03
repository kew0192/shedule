package users

import (
	"context"
	"errors"
	"log"
	"math/rand/v2"
)

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func GenerateCode() string {

	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	bytes := make([]byte, 6)
	for i := range bytes {
		bytes[i] = charset[rand.IntN(len(charset))]
	}
	return string(bytes)
}

func (s *Service) CreateUser(ctx context.Context, user *User) (*User, error) {
	if user.First_name == "" {
		return nil, errors.New("First_name required")
	}
	if user.Last_name == "" {
		return nil, errors.New("Last_name required")
	}
	if user.Role == "" {
		return nil, errors.New("Role required")
	}
	Code := GenerateCode()
	//hashed, _ := bcrypt.GenerateFromPassword([]byte(Code), 10)
	user.Code = Code
	log.Printf("📌 Сгенерирован code: %s", user.Code)
	return user, nil
}

func (s *Service) DeleteUser(ctx context.Context, id int) error {
	return s.Repo.DeleteUserDB(ctx, id)
}

func (s *Service) FindUser(ctx context.Context, id int) (*User, error) {
	user, err := s.Repo.FindUserDB(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	user.Code = ""

	return user, nil
}
