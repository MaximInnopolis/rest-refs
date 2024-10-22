package postgresql

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
	"rest-refs/internal/app/models"
	"rest-refs/internal/app/repository/database"
)

var ErrUserNotFound = errors.New("пользователь не найден")

// UserPostgres implements the UserRepo interface for PostgreSQL database operations related to users
type UserPostgres struct {
	db     database.Database
	logger *logrus.Logger
}

// NewUserPostgres creates new UserPostgres instance with provided database connection and logger
func NewUserPostgres(db database.Database, logger *logrus.Logger) *UserPostgres {
	return &UserPostgres{
		db:     db,
		logger: logger,
	}
}

// GetByEmail retrieves user from the users table by email and verifies provided password
// Returns user and error if user is not found
func (up *UserPostgres) GetByEmail(email string) (models.User, error) {
	up.logger.Debugf("GetByEmail[repo]: Получение пользователя по email: %s", email)

	query := `SELECT id, email, password, created_at FROM users WHERE email = $1`
	var dbUser models.User
	ctx := context.Background()

	// Create context with timeout to cancel query execution if it takes too long
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel() // Cancel context after function ends

	// Use a channel to get user from goroutine
	userChan := make(chan models.User)

	// Use a channel to get error from goroutine
	errChan := make(chan error)

	go func() {
		// Begin transaction
		tx, err := up.db.GetPool().BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			up.logger.Errorf("GetByEmail[repo]: Ошибка начала транзакции: %s", err)
			errChan <- err
			return
		}
		defer tx.Rollback(ctx) // Rollback transaction if function returns error

		// Execute query and scan returned user into dbUser object
		err = tx.QueryRow(ctx, query, email).
			Scan(&dbUser.ID, &dbUser.Email, &dbUser.Password, &dbUser.CreatedAt)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				up.logger.Errorf("GetByEmail[repo]: Пользователь по email: %s не найден", email)
				errChan <- ErrUserNotFound
				return
			}
			errChan <- err
			return
		}

		// Commit transaction
		if err = tx.Commit(ctx); err != nil {
			up.logger.Errorf("GetByEmail[repo]: Ошибка при коммите транзакции: %s", err)
			errChan <- err
			return
		}

		userChan <- dbUser
	}()

	select {
	case user := <-userChan:
		up.logger.Infof("GetByEmail[repo]: Пользователь успешно получен по email: %s", email)
		return user, nil
	case err := <-errChan:
		return models.User{}, err
	case <-ctx.Done():
		up.logger.Errorf("GetByEmail[repo]: Время ожидания превышено для пользователя: %s", email)
		return models.User{}, ctx.Err()
	}
}

// Create inserts new user into the users table and returns error if operation fails
func (up *UserPostgres) Create(user models.User) error {
	up.logger.Debugf("Create[repo]: Создание нового пользователя: %s", user.Email)

	query := `INSERT INTO users (email, password, created_at) 
	          VALUES ($1, $2, NOW()) RETURNING id, created_at`
	ctx := context.Background()

	// Create context with timeout to cancel query execution if it takes too long
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel() // Cancel context after function ends

	// Use a channel to get error from goroutine
	errChan := make(chan error)

	go func() {
		// Begin transaction
		tx, err := up.db.GetPool().BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			up.logger.Errorf("Create[repo]: Ошибка начала транзакции для пользователя: %s, ошибка: %s", user.Email, err)
			errChan <- err
			return
		}
		defer tx.Rollback(ctx) // Rollback transaction if function returns error

		// Execute query and scan returned ID and created_at into user object
		err = tx.QueryRow(ctx, query, user.Email, user.Password).
			Scan(&user.ID, &user.CreatedAt)
		if err != nil {
			up.logger.Errorf("Create[repo]: Ошибка создания пользователя: %s, ошибка: %s", user.Email, err)
			errChan <- err
			return
		}

		// Commit transaction
		if err = tx.Commit(ctx); err != nil {
			up.logger.Errorf("Create[repo]: Ошибка коммита транзакции для пользователя: %s, ошибка: %s", user.Email, err)
			errChan <- err
			return
		}

		up.logger.Infof("Create[repo]: Новый пользователь: %s успешно создан", user.Email)
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		up.logger.Errorf("Create[repo]: Время ожидания превышено для пользователя: %s", user.Email)
		return ctx.Err()
	}
}
