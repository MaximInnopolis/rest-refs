package postgresql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/sirupsen/logrus"
	"rest-refs/internal/app/models"
	"rest-refs/internal/app/repository/database"
)

var ErrReferralNotFound = errors.New("реферал не найден")

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

func (r *ReferralPostgres) GetReferralsByReferrerID(referrerID int) ([]models.Referral, error) {
	r.logger.Debugf("GetReferralsByReferrerID[repo]: Получение рефералов для реферера с id: %d", referrerID)

	query := `SELECT id, email, referral_code_id, referrer_id, created_at FROM referrals WHERE referrer_id = $1;`
	ctx := context.Background()

	rows, err := r.db.GetPool().Query(ctx, query, referrerID)
	if err != nil {
		r.logger.Errorf("GetReferralsByReferrerID[repo]: Ошибка при выполнении запроса: %s", err)
		return nil, err
	}
	defer rows.Close()

	var referrals []models.Referral
	var referralCodeID sql.NullInt32
	for rows.Next() {
		var referral models.Referral
		err = rows.Scan(&referral.ID, &referral.Email, &referralCodeID,
			&referral.ReferrerID, &referral.CreatedAt)
		if err != nil {
			r.logger.Errorf("GetReferralsByReferrerID[repo]: Ошибка сканировании строки: %s", err)
			return nil, err
		}

		// Check if referralCode is Null
		if referralCodeID.Valid {
			referral.ReferralCodeID = int(referralCodeID.Int32)
		} else {
			referral.ReferralCodeID = 0
		}

		referrals = append(referrals, referral)
	}

	if rows.Err() != nil {
		r.logger.Errorf("GetReferralsByReferrerID[repo]: Ошибка после итерации по строкам: %s", err)
		return nil, err
	}

	r.logger.Infof("GetReferralsByReferrerID[repo]: Рефералы успешно получены для реферера с id: %d", referrerID)
	return referrals, nil
}

func (r *ReferralPostgres) Create(referral models.Referral) error {
	r.logger.Debugf("Create[repo]: Создание нового реферала")

	query := `INSERT INTO referrals (email, referral_code_id, referrer_id, created_at) 
	          VALUES ($1, $2, $3, NOW()) RETURNING id, created_at`
	ctx := context.Background()

	// Execute query and scan returned ID, created_at, and updated_at into referral object
	err := r.db.GetPool().QueryRow(ctx, query, referral.Email, referral.ReferralCodeID,
		referral.ReferrerID).Scan(&referral.ID, &referral.CreatedAt)
	if err != nil {
		r.logger.Errorf("Create[repo]: Ошибка создания реферала: %s", err)
		return err
	}

	r.logger.Infof("Create[repo]: Новый реферал успешно создан")
	return nil
}
