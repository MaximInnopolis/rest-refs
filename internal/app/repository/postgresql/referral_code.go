package postgresql

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
	"rest-refs/internal/app/models"
	"rest-refs/internal/app/repository/database"
)

var ErrReferralCodeNotFound = errors.New("реферальный код не найден")

type ReferralCodePostgres struct {
	db             database.Database
	logger         *logrus.Logger
	getterPostgres *UserGetterPostgres
}

// NewReferralCodePostgres creates new ReferralCodePostgres instance with provided database connection and logger
func NewReferralCodePostgres(db database.Database, getterPostgres *UserGetterPostgres, logger *logrus.Logger) *ReferralCodePostgres {
	return &ReferralCodePostgres{
		db:             db,
		getterPostgres: getterPostgres,
		logger:         logger,
	}
}

func (r *ReferralCodePostgres) GetByEmail(email string) (models.User, error) {
	return r.getterPostgres.GetByEmail(email)
}

func (r *ReferralCodePostgres) Create(referralCode models.ReferralCode) (models.ReferralCode, error) {
	r.logger.Debugf("Create[repo]: Создание нового реферального кода для пользователя с id: %d", referralCode.ReferrerID)

	query := `INSERT INTO referral_codes (code, expires_at, referrer_id, created_at, updated_at)
              VALUES ($1, $2, $3, NOW(), NOW()) 
              RETURNING id, created_at, updated_at;`
	ctx := context.Background()

	err := r.db.GetPool().QueryRow(ctx, query, referralCode.Code, referralCode.Expiration, referralCode.ReferrerID).Scan(
		&referralCode.ID,
		&referralCode.CreatedAt,
		&referralCode.UpdatedAt,
	)
	if err != nil {
		r.logger.Errorf("Create[repo]: Ошибка создания реферального кода: %+v в базе: %s", referralCode, err)
		return models.ReferralCode{}, err
	}

	r.logger.Infof("Create[repo]: Новый реферальный код для пользователя c id: %d успешно создан", referralCode.ReferrerID)
	return referralCode, nil
}

// DeleteActiveReferralCodeByID removes referral code from database by ID
// If referral code with given id not found, returns ErrReferralCodeNotFound
func (r *ReferralCodePostgres) DeleteActiveReferralCodeByID(id int) error {
	r.logger.Debugf("DeleteActiveReferralCodeByID[repo]: Удаление реферального кода с id: %d", id)

	query := `DELETE FROM referral_codes WHERE id = $1`
	ctx := context.Background()

	// Execute delete query and check how many rows were affected
	result, err := r.db.GetPool().Exec(ctx, query, id)
	if err != nil {
		r.logger.Errorf("DeleteActiveReferralCodeByID[repository]: Ошибка удаления реферального кода с id: %d: %s", id, err)
		return err
	}

	// If no rows affected, return ErrReferralCodeNotFound
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warnf("DeleteActiveReferralCodeByID[repo]: Реферальный код с id: %d не найден для удаления", id)
		return ErrReferralCodeNotFound
	}

	r.logger.Infof("DeleteActiveReferralCodeByID[repository]: Реферальный код с id %d успешно удален", id)
	return nil
}

func (r *ReferralCodePostgres) GetActiveReferralCodeByUserID(referrerID int) (models.ReferralCode, error) {
	r.logger.Debugf("GetActiveReferralCodeByUserID[repo]: Получение активного реферального кода для рефера с id: %d", referrerID)

	query := ` SELECT id, code, expires_at, referrer_id, created_at, updated_at
 			   FROM referral_codes WHERE referrer_id = $1 AND expires_at > NOW()
 			   LIMIT 1;`
	var referralCode models.ReferralCode
	ctx := context.Background()

	err := r.db.GetPool().QueryRow(ctx, query, referrerID).Scan(
		&referralCode.ID,
		&referralCode.Code,
		&referralCode.Expiration,
		&referralCode.ReferrerID,
		&referralCode.CreatedAt,
		&referralCode.UpdatedAt,
	)

	if err != nil {
		// If no rows returned, return ErrReferralCodeNotFound.
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Errorf("GetActiveReferralCodeByUserID[repo]: Активного реферального кода для рефера с id: %d не найдено", referrerID)
			return models.ReferralCode{}, ErrReferralCodeNotFound
		}
		r.logger.Errorf("GetActiveReferralCodeByUserID[repo]: Ошибка при получении реферального кода для рефера с id: %d: %s", referrerID, err)
		return models.ReferralCode{}, err
	}

	r.logger.Infof("GetActiveReferralCodeByUserID[repo]: Найден активный реферальный код для реферера с id: %d", referrerID)
	return referralCode, nil
}
