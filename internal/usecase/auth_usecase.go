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

func (u *AuthUsecase) Register(idToken, name string) error {
	client, err := config.App.Auth(context.Background())
	if err != nil {
		return errors.New("firebase auth init failed")
	}

	token, err := client.VerifyIDToken(context.Background(), idToken)
	if err != nil {
		return errors.New("invalid firebase token")
	}

	emailVerified, ok := token.Claims["email_verified"].(bool)
	if !ok || !emailVerified {
		return errors.New("email not verified")
	}

	email, _ := token.Claims["email"].(string)

	existing, _ := u.userRepo.FindByUID(token.UID)
	if existing != nil {
		return errors.New("user already registered")
	}

	newUser := &domain.User{
		FirebaseUID: token.UID,
		Email:       email,
		Name:        name,
	}
	return u.userRepo.Create(newUser)
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
		return "", errors.New("user not found, please register first")
	}

	// Update email jika berubah di Firebase
	if user.Email != email {
		u.userRepo.UpdateEmail(user.ID, email)
		user.Email = email
	}

	jwt := config.GenerateJWT(user.ID, user.Email)
	return jwt, nil
}
