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

func (u *AuthUsecase) Login(idToken string) (string, error) {
	client, _ := config.App.Auth(context.Background())

	token, err := client.VerifyIDToken(context.Background(), idToken)
	if err != nil {
		return "", err
	}

	if !token.Claims["email_verified"].(bool) {
		return "", errors.New("email not verified")
	}

	email := token.Claims["email"].(string)

	user, _ := u.userRepo.FindByUID(token.UID)

	if user == nil {
		newUser := &domain.User{
			FirebaseUID: token.UID,
			Email:       email,
		}
		u.userRepo.Create(newUser)
		user = newUser
	}

	jwt := config.GenerateJWT(user.ID, user.Email)

	return jwt, nil
}
