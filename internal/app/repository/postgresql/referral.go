package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
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

	// Create context with timeout to cancel query execution if it takes too long
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel() // Cancel context after function ends

	// Use a channel to get error from goroutine
	errChan := make(chan error)

	// Use a channel to get referrals from goroutine
	referralsChan := make(chan []models.Referral)

	go func() {
		// Begin transaction
		tx, err := r.db.GetPool().BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			r.logger.Errorf("GetReferralsByReferrerID[repo]: Ошибка начала транзакции: %s", err)
			errChan <- err
			return
		}
		defer tx.Rollback(ctx) // Rollback transaction if function returns error

		rows, err := tx.Query(ctx, query, referrerID)
		if err != nil {
			r.logger.Errorf("GetReferralsByReferrerID[repo]: Ошибка при выполнении запроса: %s", err)
			errChan <- err
			return
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
				errChan <- err
				return
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
			errChan <- err
			return
		}

		// Commit transaction
		if err = tx.Commit(ctx); err != nil {
			r.logger.Errorf("GetReferralsByReferrerID[repo]: Ошибка коммита транзакции: %s", err)
			errChan <- err
			return
		}

		referralsChan <- referrals
	}()

	select {
	case referrals := <-referralsChan:
		r.logger.Infof("GetReferralsByReferrerID[repo]: Рефералы успешно получены для реферера с id: %d", referrerID)
		return referrals, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		r.logger.Errorf("GetReferralsByReferrerID[repo]: Время ожидания превышено для реферера с id: %d", referrerID)
		return nil, ctx.Err()
	}
}

func (r *ReferralPostgres) Create(referral models.Referral) error {
	r.logger.Debugf("Create[repo]: Создание нового реферала")

	query := `INSERT INTO referrals (email, referral_code_id, referrer_id, created_at) 
	          VALUES ($1, $2, $3, NOW()) RETURNING id, created_at`
	ctx := context.Background()

	// Create context with timeout to cancel query execution if it takes too long
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel() // Cancel context after function ends

	// Use a channel to get error from goroutine
	errChan := make(chan error)

	go func() {
		// Begin transaction
		tx, err := r.db.GetPool().BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			r.logger.Errorf("Create[repo]: Ошибка начала транзакции для реферала: %s, ошибка: %s", referral.Email, err)
			errChan <- err
			return
		}
		defer tx.Rollback(ctx) // Rollback transaction if function returns error

		// Execute query and scan returned ID, created_at, and updated_at into referral object
		err = tx.QueryRow(ctx, query, referral.Email, referral.ReferralCodeID,
			referral.ReferrerID).Scan(&referral.ID, &referral.CreatedAt)
		if err != nil {
			r.logger.Errorf("Create[repo]: Ошибка создания реферала: %s", err)
			errChan <- err
			return
		}

		// Commit transaction
		if err = tx.Commit(ctx); err != nil {
			r.logger.Errorf("Create[repo]: Ошибка коммита транзакции для реферала: %s, ошибка: %s", referral.Email, err)
			errChan <- err
			return
		}

		r.logger.Infof("Create[repo]: Новый реферал успешно создан")
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		r.logger.Errorf("Create[repo]: Время ожидания превышено для пользователя: %s", referral.Email)
		return ctx.Err()
	}
}
