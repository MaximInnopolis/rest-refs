package api

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"rest-refs/internal/app/models"
	"rest-refs/internal/app/repository"
)

// Authorization defines methods related to user authorization and token management
type Authorization interface {
	RegisterUser(user models.User) error
	GetUserByEmail(email string) (models.User, error)
	GenerateToken(user models.User) (string, error)
	IsTokenValid(tokenString string) (bool, jwt.MapClaims, error)
}

// ReferralCode defines methods for handling referral codes
type ReferralCode interface {
	CreateReferralCode(referralCode models.ReferralCode) (models.ReferralCode, error)
	DeleteReferralCode(referrerID int) error
	GetReferralCodeByReferrerEmail(email string) (models.ReferralCode, error)
	GetIDByReferralCode(code string) (int, error)
	GetReferrerIDByReferralCode(code string) (int, error)
}

// Referral defines methods related to referral management
type Referral interface {
	GetReferralsByReferrerID(referrerID int) ([]models.ReferralInfoResponse, error)
	RegisterWithReferralCode(referralCode string, user models.User) error
}

// Service aggregates different services related to user authorization, referral codes, and referrals
type Service struct {
	Authorization
	Referral
	ReferralCode
}

// New returns new instance of Service, initializing dependencies
// It takes repository that holds database access logic
func New(repo *repository.Repository, logger *logrus.Logger) *Service {
	authService := NewAuthService(repo.UserRepo, logger)
	referralCodeService := NewReferralCodeService(repo.ReferralCodeRepo, authService, logger)
	referralService := NewReferralService(repo.ReferralRepo, referralCodeService, logger)

	return &Service{
		Authorization: authService,
		ReferralCode:  referralCodeService,
		Referral:      referralService,
	}
}
