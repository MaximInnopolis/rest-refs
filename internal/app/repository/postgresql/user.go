package postgresql

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"rest-refs/internal/app/models"
	"rest-refs/internal/app/repository/database"
)

var ErrNotFound = errors.New("user not found")

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

// Create inserts new user into the users table and returns error if operation fails
func (up *UserPostgres) Create(user models.User) error {
	up.logger.Infof("Create[repo]: Создание нового пользователя: %+v", user)

	query := `INSERT INTO users (email, password, created_at) 
	          VALUES ($1, $2, NOW()) RETURNING id, created_at`
	ctx := context.Background()

	// Execute query and scan returned ID, created_at, and updated_at into user object
	err := up.db.GetPool().QueryRow(ctx, query, user.Email, user.Password).
		Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		up.logger.Errorf("Create[repo]: Ошибка создания пользователя: %+v, ошибка: %s", user, err)
		return err
	}
	return nil
}

// Get retrieves user from the users table by email and verifies provided password
// Returns user and error if user is not found or if password is incorrect
func (up *UserPostgres) Get(user models.User) (models.User, error) {
	up.logger.Infof("Get[repo]: Получение пользователя: %+v", user)

	query := `SELECT id, email, password, created_at FROM users WHERE email = $1`
	var dbUser models.User

	ctx := context.Background()

	err := up.db.GetPool().QueryRow(ctx, query, user.Email).
		Scan(&dbUser.ID, &dbUser.Email, &dbUser.Password, &dbUser.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrNotFound
		}
		return models.User{}, err
	}

	// Compare provided password with hashed password stored in database
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password))
	if err != nil {
		return models.User{}, errors.New("invalid password")
	}

	return dbUser, nil
}
