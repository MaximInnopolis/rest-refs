package api

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"rest-refs/internal/app/models"
	"rest-refs/internal/app/repository"
)

type Authorization interface {
	CreateUser(user models.User) error
	GenerateToken(user models.User) (string, error)
	IsTokenValid(tokenString string) (bool, jwt.MapClaims, error)
}

type Refferal interface {
	CreateReferralCode(referralCode models.ReferralCode) (models.ReferralCode, error)
	DeleteReferralCode(referrerID int) error
	GetReferralCodeByReferrerEmail(email string) (models.ReferralCode, error)

	GetReferralsByReferrerID(referrerID int) ([]models.ReferralInfoResponse, error)

	RegisterWithReferralCode(referralCode string, user models.User) error
}

type Service struct {
	Authorization
	Refferal
}

// New returns new instance of Service, initializing dependencies
// It takes repository that holds database access logic
func New(repo *repository.Repository, logger *logrus.Logger) *Service {
	return &Service{
		Authorization: NewAuthService(repo.UserRepo, logger),
		Refferal:      NewReferralService(repo.ReferralRepo, logger),
	}
}
