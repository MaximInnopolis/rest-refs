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

func (r *ReferralPostgres) Create(referralCode models.ReferralCode) (models.ReferralCode, error) {
	r.logger.Debugf("Create[repo]: Создание нового реферального кода для пользователя %d", referralCode.ReferrerID)

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

	r.logger.Infof("Create[repo]: Новый реферальный код для пользователя %d успешно создан", referralCode.ReferrerID)
	return referralCode, nil
}

// Delete removes referral code from database by ID
// If referral code with given referrerID not found, returns ErrReferralCodeNotFound
func (r *ReferralPostgres) Delete(referrerID int) error {
	r.logger.Debugf("Delete[repo]: Удаление реферального кода по id: %d", referrerID)

	query := `DELETE FROM referral_codes WHERE referrer_id = $1`
	ctx := context.Background()

	// Execute delete query and check how many rows were affected
	result, err := r.db.GetPool().Exec(ctx, query, referrerID)
	if err != nil {
		r.logger.Errorf("Delete[repository]: Ошибка удаления реферального кода пользователя с id %d: %s", referrerID, err)
		return err
	}

	// If no rows affected, return ErrReferralCodeNotFound
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warnf("Delete[repo]: Реферальный код пользователя с id %d не найден для удаления", referrerID)
		return ErrReferralCodeNotFound
	}

	r.logger.Infof("Delete[repository]: Реферальный код пользователя с id %d успешно удален", referrerID)
	return nil
}

func (r *ReferralPostgres) GetReferrerIDByEmail(email string) (int, error) {
	r.logger.Debugf("GetReferrerIDByEmail[repo]: Получение id реферера по email: %s", email)

	query := `SELECT id FROM users WHERE email = $1`
	var dbID int
	ctx := context.Background()

	err := r.db.GetPool().QueryRow(ctx, query, email).
		Scan(&dbID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Errorf("GetReferrerIDByEmail[repo]: Id пользователя по email %s не найден", email)
			return 0, ErrUserNotFound
		}
		return 0, err
	}

	r.logger.Infof("GetReferrerIDByEmail[repo]: Id реферера успешно получен по email: %s", email)
	return dbID, nil
}

func (r *ReferralPostgres) GetActiveReferralCodeByReferrerID(referrerID int) (models.ReferralCode, error) {
	r.logger.Debugf("GetActiveReferralCodeByReferrerID[repo]: Получение активного реферального кода для рефера с ID: %d", referrerID)

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
			r.logger.Errorf("GetActiveReferralCodeByReferrerID[repo]: Активного реферального кода для рефера с ID %d не найдено", referrerID)
			return models.ReferralCode{}, ErrReferralCodeNotFound
		}
		r.logger.Errorf("GetActiveReferralCodeByReferrerID[repo]: Ошибка при получении реферального кода: %s", err)
		return models.ReferralCode{}, err
	}

	r.logger.Infof("GetActiveReferralCodeByReferrerID[repo]: Найден активный реферальный код для реферера с ID: %d", referrerID)
	return referralCode, nil
}

func (r *ReferralPostgres) GetReferralsByReferrerID(id int) ([]models.Referral, error) {
	r.logger.Debugf("GetReferralsByReferrerID[repo]: Получение рефералов для реферера с id %d", referrerID)

	query := `SELECT id, email, referral_code_id, referrer_id, created_at FROM referrals WHERE referrer_id = $1;`
	ctx := context.Background()

	rows, err := r.db.GetPool().Query(ctx, query, id)
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

	r.logger.Infof("GetReferralsByReferrerID[repo]: Рефералы успешно получены для реферера с id %d", referrerID)
	return referrals, nil
}

//func (r *ReferralPostgres) GetByEmail(email string) (models.ReferralCode, error) {
//
//}
//
//func (r *ReferralPostgres) Get(code string) (models.ReferralCode, error)      {}

func (r *ReferralPostgres) Register(referralCodeID, referredUserID int) error {}
