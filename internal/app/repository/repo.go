package repository

import (
	"github.com/sirupsen/logrus"
	"rest-refs/internal/app/models"
	"rest-refs/internal/app/repository/database"
	"rest-refs/internal/app/repository/postgresql"
)

// UserRepo defines interface for user-related database operations
type UserRepo interface {
	Create(user models.User) error
	Get(user models.User) (models.User, error)
}

// ReferralRepo defines interface for referral-related database operations
type ReferralRepo interface {
	Create(referralCode models.ReferralCode) (models.ReferralCode, error)
	Delete(referrerID int) error
	GetByEmail(email string) (models.ReferralCode, error)
	Get(code string) (models.ReferralCode, error)
	Register(referralCodeID, referredUserID int) error
	GetByReferrerID(referrerID int) ([]models.ReferralInfoResponse, error)
	GetActiveReferralCodeByReferrerID(referrerID int) (*models.ReferralCode, error)
}

// Repository combines UserRepo and ReferralRepo interfaces into single struct
type Repository struct {
	UserRepo
	ReferralRepo
}

// New initializes and returns new Repository instance with PostgreSQL implementations for UserRepo and ReferralRepo
func New(db database.Database, logger *logrus.Logger) *Repository {
	return &Repository{
		UserRepo:     postgresql.NewUserPostgres(db, logger),
		ReferralRepo: postgresql.NewReferralPostgres(db, logger),
	}
}
