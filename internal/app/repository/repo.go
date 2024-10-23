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
	GetByEmail(email string) (models.User, error)
}

// ReferralCodeRepo defines interface for referral code-related database operations
type ReferralCodeRepo interface {
	Create(referralCode models.ReferralCode) (models.ReferralCode, error)
	DeleteActiveReferralCodeByID(id int) error
	GetActiveReferralCodeByUserID(referrerID int) (models.ReferralCode, error)
	GetIDByReferralCode(code string) (int, error)
	GetReferrerIDByReferralCode(code string) (int, error)
}

// ReferralRepo defines interface for referral-related database operations
type ReferralRepo interface {
	GetReferralsByReferrerID(id int) ([]models.Referral, error)
	Create(referral models.Referral) error
}

// Repository combines UserRepo, ReferralCodeRepo, and ReferralRepo interfaces into single struct
type Repository struct {
	UserRepo
	ReferralRepo
	ReferralCodeRepo
}

// New initializes and returns new Repository instance with PostgreSQL implementations for UserRepo and ReferralRepo
func New(db database.Database, logger *logrus.Logger) *Repository {

	return &Repository{
		UserRepo:         postgresql.NewUserPostgres(db, logger),
		ReferralCodeRepo: postgresql.NewReferralCodePostgres(db, logger),
		ReferralRepo:     postgresql.NewReferralPostgres(db, logger),
	}
}
