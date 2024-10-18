package postgresql

import (
	"github.com/sirupsen/logrus"
	"rest-refs/internal/app/models"
	"rest-refs/internal/app/repository/database"
)

type ReferralPostgres struct {
	db     database.Database
	logger *logrus.Logger
}

// NewReferralPostgres creates new ReferralPostgres instance with provided database connection and logger
func NewReferralPostgres(db database.Database, logger *logrus.Logger) *ReferralPostgres {
	return &ReferralPostgres{
		db:     db,
		logger: logger,
	}
}

func (r *ReferralPostgres) Create(referralCode models.ReferralCode) (models.ReferralCode, error)  {}
func (r *ReferralPostgres) Delete(referrerID int) error                                           {}
func (r *ReferralPostgres) GetByEmail(email string) (models.ReferralCode, error)                  {}
func (r *ReferralPostgres) Get(code string) (models.ReferralCode, error)                          {}
func (r *ReferralPostgres) Register(referralCodeID, referredUserID int) error                     {}
func (r *ReferralPostgres) GetByReferrerID(referrerID int) ([]models.ReferralInfoResponse, error) {}
func (r *ReferralPostgres) GetActiveReferralCodeByReferrerID(referrerID int) (*models.ReferralCode, error) {
}
