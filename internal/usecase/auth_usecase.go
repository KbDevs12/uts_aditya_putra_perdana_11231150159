package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"regexp"
	"strings"
	"time"

	"backend/config"
	"backend/internal/domain"
	"backend/internal/repository"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/redis/go-redis/v9"
)

type AuthUsecase struct {
	userRepo    *repository.UserRepo
	redisClient *redis.Client
}

type FirebaseIdentity struct {
	UID           string
	Email         string
	EmailVerified bool
}

const otpKeyPrefix = "wallet:email_otp:"

func NewAuthUsecase(userRepo *repository.UserRepo, redisClient *redis.Client) *AuthUsecase {
	return &AuthUsecase{userRepo: userRepo, redisClient: redisClient}
}

func (u *AuthUsecase) Register(idToken, name string) error {
	identity, err := verifyFirebaseToken(idToken)
	if err != nil {
		return err
	}
	if name == "" {
		name = strings.Split(identity.Email, "@")[0]
	}

	existing, err := u.userRepo.FindByUID(identity.UID)
	if existing != nil && err == nil {
		existing.Email = identity.Email
		existing.Name = name
		if identity.EmailVerified {
			existing.EmailVerified = true
		}
		if err := u.userRepo.Save(existing); err != nil {
			return err
		}
		return u.SendEmailOTP(identity.Email)
	}

	newUser := &domain.User{
		FirebaseUID:      identity.UID,
		Email:            identity.Email,
		Name:             name,
		EmailVerified:    identity.EmailVerified,
		TwoFactorMethod:  "smtp",
		TwoFactorEnabled: false,
	}
	if err := u.userRepo.Create(newUser); err != nil {
		return err
	}
	return u.SendEmailOTP(identity.Email)
}

func (u *AuthUsecase) Login(idToken string) (string, bool, string, error) {
	identity, err := verifyFirebaseToken(idToken)
	if err != nil {
		return "", false, "", err
	}

	requireFirebaseEmailVerified := strings.EqualFold(os.Getenv("REQUIRE_FIREBASE_EMAIL_VERIFIED"), "true")
	if requireFirebaseEmailVerified && !identity.EmailVerified {
		return "", false, "", errors.New("email not verified")
	}

	user, _ := u.userRepo.FindByUID(identity.UID)
	if user == nil {
		return "", false, "", errors.New("user not found, please register first")
	}

	if user.Email != identity.Email {
		_ = u.userRepo.UpdateEmail(user.ID, identity.Email)
		user.Email = identity.Email
	}
	if identity.EmailVerified && !user.EmailVerified {
		_ = u.userRepo.MarkEmailVerified(user.ID)
		user.EmailVerified = true
	}

	now := time.Now()
	_ = u.userRepo.TouchLastLogin(user.ID, now)
	return config.GenerateJWT(user.ID, user.Email), user.TwoFactorEnabled, user.TwoFactorMethod, nil
}

func (u *AuthUsecase) SendEmailOTP(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if !isEmail(email) {
		return errors.New("valid email is required")
	}

	code, err := randomOTP()
	if err != nil {
		return err
	}

	if u.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := u.redisClient.Set(ctx, otpKey(email), code, 10*time.Minute).Err(); err != nil {
			return err
		}
	} else {
		// Fallback untuk MVP lokal kalau Redis belum dinyalakan.
		log.Printf("Redis disabled, OTP not persisted. Email=%s Code=%s", email, code)
	}

	return config.SendOTPEmail(email, code)
}

func (u *AuthUsecase) VerifyEmailOTP(email, code string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	code = strings.TrimSpace(code)
	if !isEmail(email) || len(code) != 6 {
		return "", errors.New("email and 6 digit code are required")
	}

	if err := u.verifyStoredEmailOTP(email, code); err != nil {
		return "", err
	}

	user, err := u.userRepo.FindByEmail(email)
	if err != nil || user == nil {
		return "", errors.New("user not found")
	}
	if err := u.userRepo.MarkEmailVerified(user.ID); err != nil {
		return "", err
	}
	return config.GenerateJWT(user.ID, user.Email), nil
}

func (u *AuthUsecase) SetupTwoFactor(userID int64, method string) (map[string]string, error) {
	method = strings.TrimSpace(strings.ToLower(method))
	if method == "" {
		method = "smtp"
	}
	if method != "smtp" && method != "totp" && method != "notif" {
		return nil, errors.New("2fa method must be smtp, totp, or notif")
	}

	user, err := u.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	response := map[string]string{
		"method": method,
	}

	switch method {
	case "smtp":
		if err := u.SendEmailOTP(user.Email); err != nil {
			return nil, err
		}
		response["message"] = "OTP 6 digit sudah dikirim ke email. Masukkan kode untuk mengaktifkan 2FA Email."
		return response, nil
	case "totp":
		accountName := user.Email
		if accountName == "" {
			accountName = fmt.Sprintf("user-%d", user.ID)
		}
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "Kantongin",
			AccountName: accountName,
			Period:      30,
			SecretSize:  20,
			Secret:      nil,
			Digits:      otp.DigitsSix,
			Algorithm:   otp.AlgorithmSHA1,
		})
		if err != nil {
			return nil, err
		}
		secret := key.Secret()
		if err := u.userRepo.SaveTOTPSecret(userID, secret); err != nil {
			return nil, err
		}
		response["secret"] = secret
		response["otpauth_url"] = key.URL()
		response["issuer"] = "Kantongin"
		response["account_name"] = accountName
		response["period_seconds"] = "30"
		response["message"] = "Scan QR di Google Authenticator, lalu masukkan kode 6 digit yang muncul."
		return response, nil
	case "notif":
		// Untuk MVP, push notification tetap berupa flow approval sederhana.
		// FCM token tetap disimpan lewat endpoint /auth/fcm-token.
		response["message"] = "Simulasi push notification aktif. Masukkan kode 000000 untuk approve demo."
		return response, nil
	default:
		return nil, errors.New("unsupported 2fa method")
	}
}

func (u *AuthUsecase) VerifyTwoFactor(userID int64, method, code string) error {
	method = strings.TrimSpace(strings.ToLower(method))
	code = strings.TrimSpace(code)
	user, err := u.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}
	if method == "" {
		method = user.TwoFactorMethod
	}

	switch method {
	case "smtp":
		if len(code) != 6 {
			return errors.New("6 digit code is required")
		}
		if err := u.verifyStoredEmailOTP(user.Email, code); err != nil {
			return err
		}
		return u.userRepo.UpdateTwoFactor(userID, "smtp", true)
	case "totp":
		if len(code) != 6 {
			return errors.New("6 digit code is required")
		}
		if strings.TrimSpace(user.TOTPSecret) == "" {
			return errors.New("totp has not been set up")
		}
		if !totp.Validate(code, user.TOTPSecret) {
			return errors.New("invalid authenticator code")
		}
		return u.userRepo.MarkTOTPVerified(userID, time.Now())
	case "notif":
		if code != "000000" && !strings.EqualFold(code, "approved") {
			return errors.New("push approval is required")
		}
		return u.userRepo.UpdateTwoFactor(userID, "notif", true)
	default:
		return errors.New("2fa method must be smtp, totp, or notif")
	}
}

func (u *AuthUsecase) SaveFCMToken(userID int64, token string) error {
	if strings.TrimSpace(token) == "" {
		return errors.New("fcm token is required")
	}
	return u.userRepo.UpdateFCMToken(userID, token)
}

func (u *AuthUsecase) verifyStoredEmailOTP(email, code string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	code = strings.TrimSpace(code)

	if u.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		stored, err := u.redisClient.Get(ctx, otpKey(email)).Result()
		if err != nil || stored != code {
			return errors.New("invalid or expired otp")
		}
		_ = u.redisClient.Del(ctx, otpKey(email)).Err()
		return nil
	}

	if os.Getenv("ALLOW_MOCK_OTP") == "true" && code == "123456" {
		return nil
	}
	return errors.New("redis is required for otp verification")
}

func verifyFirebaseToken(idToken string) (*FirebaseIdentity, error) {
	idToken = strings.TrimSpace(idToken)
	if idToken == "" {
		return nil, errors.New("invalid firebase token")
	}

	if config.App == nil {
		if !strings.EqualFold(os.Getenv("ALLOW_MOCK_AUTH"), "true") {
			return nil, errors.New("firebase auth init failed")
		}
		email := idToken
		if !isEmail(email) {
			email = "demo@kantongin.local"
		}
		hash := sha256.Sum256([]byte(email))
		return &FirebaseIdentity{UID: hex.EncodeToString(hash[:8]), Email: email, EmailVerified: true}, nil
	}

	client, err := config.App.Auth(context.Background())
	if err != nil {
		return nil, errors.New("firebase auth init failed")
	}

	token, err := client.VerifyIDToken(context.Background(), idToken)
	if err != nil {
		return nil, errors.New("invalid firebase token")
	}

	email, _ := token.Claims["email"].(string)
	if email == "" {
		email = token.UID + "@firebase.local"
	}
	emailVerified, _ := token.Claims["email_verified"].(bool)
	return &FirebaseIdentity{UID: token.UID, Email: strings.ToLower(email), EmailVerified: emailVerified}, nil
}

func randomOTP() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func generateHex(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func otpKey(email string) string { return otpKeyPrefix + email }

func isEmail(value string) bool {
	return regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`).MatchString(value)
}
