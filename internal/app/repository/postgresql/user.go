package postgresql

import (
	"context"
	"errors"

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

	err := up.db.GetPool().QueryRow(ctx, query, email).
		Scan(&dbUser.ID, &dbUser.Email, &dbUser.Password, &dbUser.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			up.logger.Errorf("GetByEmail[repo]: Пользователь по email: %s не найден", email)
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}

	up.logger.Infof("GetByEmail[repo]: Пользователь успешно получен по email: %s", email)
	return dbUser, nil
}

// Create inserts new user into the users table and returns error if operation fails
func (up *UserPostgres) Create(user models.User) error {
	up.logger.Debugf("Create[repo]: Создание нового пользователя: %s", user.Email)

	query := `INSERT INTO users (email, password, created_at) 
	          VALUES ($1, $2, NOW()) RETURNING id, created_at`
	ctx := context.Background()

	// Execute query and scan returned ID, created_at, and updated_at into user object
	err := up.db.GetPool().QueryRow(ctx, query, user.Email, user.Password).
		Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		up.logger.Errorf("Create[repo]: Ошибка создания пользователя: %s, ошибка: %s", user.Email, err)
		return err
	}

	up.logger.Infof("Create[repo]: Новый пользователь: %s успешно создан", user.Email)
	return nil
}
