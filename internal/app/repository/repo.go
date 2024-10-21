package repository

import (
	"github.com/sirupsen/logrus"
	"rest-refs/internal/app/models"
	"rest-refs/internal/app/repository/database"
	"rest-refs/internal/app/repository/postgresql"
)

type UserGetterRepo interface {
	GetByEmail(email string) (models.User, error)
}

// UserRepo defines interface for user-related database operations
type UserRepo interface {
	Create(user models.User) error
	UserGetterRepo
}

type ReferralCodeRepo interface {
	Create(referralCode models.ReferralCode) (models.ReferralCode, error)
	DeleteActiveReferralCodeByID(id int) error
	GetActiveReferralCodeByUserID(referrerID int) (models.ReferralCode, error)
	UserGetterRepo
}

// ReferralRepo defines interface for referral-related database operations
type ReferralRepo interface {
	GetReferralsByReferrerID(id int) ([]models.Referral, error)

	//GetByEmail(email string) (models.ReferralCode, error)
	//Get(code string) (models.ReferralCode, error)

	//Register(referralCodeID, referredUserID int) error
}

// Repository combines UserRepo and ReferralRepo interfaces into single struct
type Repository struct {
	UserRepo
	ReferralRepo
	ReferralCodeRepo
}

// New initializes and returns new Repository instance with PostgreSQL implementations for UserRepo and ReferralRepo
func New(db database.Database, logger *logrus.Logger) *Repository {
	userGetterRepo := postgresql.NewUserGetterPostgres(db, logger)

	return &Repository{
		UserRepo:         postgresql.NewUserPostgres(userGetterRepo),
		ReferralCodeRepo: postgresql.NewReferralCodePostgres(db, userGetterRepo, logger),
		ReferralRepo:     postgresql.NewReferralPostgres(db, logger),
	}
}
