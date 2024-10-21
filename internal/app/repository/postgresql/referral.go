package postgresql

import (
	"context"

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
	for rows.Next() {
		var referral models.Referral
		err = rows.Scan(&referral.ID, &referral.Email, &referral.ReferralCodeID,
			&referral.ReferrerID, &referral.CreatedAt)
		if err != nil {
			r.logger.Errorf("GetReferralsByReferrerID[repo]: Ошибка сканировании строки: %s", err)
			return nil, err
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
