package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"goapp/config"
	"goapp/models"
	"goapp/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService обрабатывает логику аутентификации/авторизации.
type AuthService struct {
	users  *repository.UserRepo
	tokens *repository.TokenRepo
	db     *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		users:  repository.NewUserRepo(db),
		tokens: repository.NewTokenRepo(db),
		db:     db,
	}
}

// Проверка Access Token
func (s *AuthService) ValidateAccessToken(tokenStr string) (*jwt.Token, error) {

	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return config.JwtAccessSecret(), nil
	})
}

// Регистрация пользователя
func (s *AuthService) Register(email, password string) (*models.User, error) {

	if _, err := s.users.FindByEmail(email); err == nil {
		return nil, errors.New("user already exists")
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:    email,
		Password: string(hash),
		Salt:     "",
	}

	if err := s.users.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Универсальная генерация токенов
func (s *AuthService) CreateTokensForUser(user *models.User) (string, string, error) {

	atClaims := jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().
			Add(config.AccessTokenExpiration()).
			Unix(),
	}

	at := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		atClaims,
	)

	accessToken, err := at.SignedString(
		config.JwtAccessSecret(),
	)

	if err != nil {
		return "", "", err
	}

	refreshPlain := uuid.New().String()

	sha := sha256.New()
	sha.Write([]byte(refreshPlain))

	rtHash := hex.EncodeToString(sha.Sum(nil))

	rt := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: rtHash,
		ExpiresAt: time.Now().
			Add(config.RefreshTokenExpiration()),
		Revoked: false,
	}

	if err := s.tokens.Save(rt); err != nil {
		return "", "", err
	}

	return accessToken, refreshPlain, nil
}

// Логин
func (s *AuthService) Login(
	email,
	password string,
) (string, string, error) {

	user, err := s.users.FindByEmail(email)

	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	if bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	) != nil {

		return "", "", errors.New("invalid credentials")
	}

	return s.CreateTokensForUser(user)
}

// Refresh токенов
func (s *AuthService) Refresh(
	refreshToken string,
) (string, string, error) {

	hashed := sha256.Sum256([]byte(refreshToken))
	tokenHash := hex.EncodeToString(hashed[:])

	rt, err := s.tokens.FindValid(tokenHash)

	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	user, err := s.users.FindByID(rt.UserID.String())

	if err != nil {
		return "", "", err
	}

	// revoke old token
	rt.Revoked = true
	_ = s.tokens.Update(rt)

	return s.CreateTokensForUser(user)
}

// Logout текущей сессии
func (s *AuthService) Logout(
	refreshToken string,
) error {

	hashed := sha256.Sum256([]byte(refreshToken))
	tokenHash := hex.EncodeToString(hashed[:])

	rt, err := s.tokens.FindValid(tokenHash)

	if err != nil {
		return err
	}

	rt.Revoked = true

	return s.tokens.Update(rt)
}

// Logout всех сессий
func (s *AuthService) LogoutAll(
	userID string,
) error {

	return s.tokens.RevokeAllByUser(userID)
}
