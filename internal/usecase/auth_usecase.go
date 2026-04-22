package usecase

import (
	"context"
	"errors"

	"backend/config"
	"backend/internal/domain"
	"backend/internal/repository"
)

type AuthUsecase struct {
	userRepo *repository.UserRepo
}

func NewAuthUsecase(userRepo *repository.UserRepo) *AuthUsecase {
	return &AuthUsecase{userRepo}
}

func (u *AuthUsecase) Login(idToken string) (string, error) {
	client, err := config.App.Auth(context.Background())
	if err != nil {
		return "", errors.New("firebase auth init failed")
	}

	token, err := client.VerifyIDToken(context.Background(), idToken)
	if err != nil {
		return "", errors.New("invalid firebase token")
	}

	emailVerified, ok := token.Claims["email_verified"].(bool)
	if !ok || !emailVerified {
		return "", errors.New("email not verified")
	}

	email, _ := token.Claims["email"].(string)

	user, _ := u.userRepo.FindByUID(token.UID)

	if user == nil {
		newUser := &domain.User{
			FirebaseUID: token.UID,
			Email:       email,
		}
		if err := u.userRepo.Create(newUser); err != nil {
			return "", errors.New("failed to create user")
		}
		user = newUser
	}

	jwt := config.GenerateJWT(user.ID, user.Email)
	return jwt, nil
}
