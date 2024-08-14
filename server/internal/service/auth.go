package service

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/egosha7/site-go/internal/authMiddleware"
	"github.com/egosha7/site-go/internal/domain"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func (s *ServiceImpl) GenerateToken(user *domain.User, ip string) (string, error) {
	s.Logger.Info("GenerateToken", zap.Any("user", user))
	id, storedPasswordHash, err := s.Repository.CheckValidUser(user.Login)
	if err != nil {
		return "", err
	}

	// Проверяем введенный пароль с сохраненным хэшем
	if !s.CheckPasswordHash(user.Password, storedPasswordHash) {
		return "", errors.New("неверный пароль")
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256, &authMiddleware.TokenClaims{
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(authMiddleware.TokenTTL).Unix(),
				IssuedAt:  time.Now().Unix(),
			},
			UserID: id,
			IP:     ip,
		},
	)

	return token.SignedString([]byte(authMiddleware.SigningKey))
}

func (s *ServiceImpl) CreateUser(login, password string) error {
	hash, err := s.HashPassword(password)
	if err != nil {
		return err
	}

	// Сохраните login и hash в базе данных
	return s.Repository.SaveUser(login, hash)
}

// HashPassword хэширует пароль с использованием bcrypt.
func (s *ServiceImpl) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPasswordHash проверяет введенный пароль на соответствие хэшу.
func (s *ServiceImpl) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
