package api

import (
	"errors"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"rest-refs/internal/app/models"
	"rest-refs/internal/app/repository"
)

var ErrUserAlreadyExists = errors.New("пользователь уже существует")

// AuthService provides authentication services using user repository
type AuthService struct {
	repo   repository.UserRepo
	logger *logrus.Logger
}

// NewAuthService creates new instance of AuthService
func NewAuthService(repo repository.UserRepo, logger *logrus.Logger) *AuthService {
	return &AuthService{
		repo:   repo,
		logger: logger,
	}
}

func (as *AuthService) GetUserByEmail(email string) (models.User, error) {
	as.logger.Debugf("GetByEmail[service]: Получение пользователя по email: %s", email)
	return as.repo.GetByEmail(email)
}

func (as *AuthService) RegisterUser(user models.User) error {
	as.logger.Debugf("RegisterUser[service]: Регистрация пользователя с email: %s", user.Email)

	_, err := as.GetUserByEmail(user.Email)
	if err == nil {
		as.logger.Errorf("RegisterUser[service]: Регистрация пользователя не удалось: " +
			"Пользователь с таким email уже существует")
		return ErrUserAlreadyExists
	}

	// Hash user's password before saving
	user.Password, err = generatePasswordHash(user.Password)
	if err != nil {
		as.logger.Errorf("RegisterUser[service]: Ошибка при хэшировании пароля: %s", err)
		return err
	}

	// Save new user in repository
	err = as.repo.Create(user)
	if err != nil {
		as.logger.Errorf("RegisterUser[service]: Ошибка при создании пользователя в базе: %s", err)
		return err
	}

	as.logger.Infof("RegisterUser[service]: Пользователь с email: %s успешно зарегистрирован", user.Email)
	return nil
}

// GenerateToken generates JWT for authenticated user
// It retrieves user from repository and creates signed JWT token
func (as *AuthService) GenerateToken(user models.User) (string, error) {
	as.logger.Debugf("GenerateToken[service]: Создание токена для пользователя: %s", user.Email)

	dbUser, err := as.repo.GetByEmail(user.Email)
	if err != nil {
		as.logger.Errorf("GenerateToken[service]: Ошибка при получении пользователя: %s для генерации токена: %s", user.Email, err)
		return "", err
	}

	// Compare provided password with hashed password stored in database
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password))
	if err != nil {
		as.logger.Errorf("GenerateToken[service]: Ошибка при сравнении паролей пользователя: %s: ,%s", user.Email, err)
		return "", errors.New("invalid password")
	}

	// Generate JWT token
	token, err := as.generateJWT(dbUser)
	if err != nil {
		as.logger.Errorf("Ошибка при генерации JWT: %s", err)
		return "", err
	}

	as.logger.Infof("GenerateToken[service]: JWT успешно сгенерирован для пользователя: %s", dbUser.Email)
	return token, nil
}

// IsTokenValid validates given JWT
// It checks token's signature, claims, and expiration time
func (as *AuthService) IsTokenValid(tokenString string) (bool, jwt.MapClaims, error) {
	as.logger.Debugf("IsTokenValid[service]: Проверка валидности токена")

	// Check token validity
	validToken, claims, err := as.checkToken(tokenString)
	if err != nil || !validToken {
		as.logger.Errorf("IsTokenValid[service]: Неверный токен: %s", err)
		return false, nil, errors.New("invalid token")
	}

	as.logger.Infof("Токен валиден")
	return true, claims, nil
}

// checkToken parses and validates JWT
// It verifies token's signature and checks expiration claim
func (as *AuthService) checkToken(tokenString string) (bool, jwt.MapClaims, error) {
	as.logger.Debugf("checkToken[service]: Проверка токена")

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(os.Getenv("SECRET_KEY")), nil
	})
	if err != nil {
		as.logger.Errorf("checkToken[service]: Ошибка при разборе токена: %s", err)
		return false, nil, err
	}

	// Check if token is valid
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		as.logger.Errorf("checkToken[service]: Некорректные claims или токен недействителен")
		return false, nil, nil
	}

	// Check if expiration claim exists and validate it
	expiration, ok := claims["exp"].(float64)
	if !ok {
		as.logger.Errorf("checkToken[service]: Некорректное время истечения токена")
		return false, nil, nil
	}

	if int64(expiration) < time.Now().Unix() {
		as.logger.Errorf("checkToken[service]: Токен истек")
		return false, nil, nil
	}

	as.logger.Infof("checkToken[service]: Токен успешно проверен")
	return true, claims, nil
}

// generateJWT generates JWT for provided user with 24-hour expiration time
func (as *AuthService) generateJWT(user models.User) (string, error) {
	as.logger.Debugf("generateJWT([service]: Генерация токена для пользователя: %s", user.Email)

	token := jwt.New(jwt.SigningMethodHS256)

	// Set standard claims
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = user.ID
	claims["sub"] = user.Email

	// Add additional claims
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	// Sign token with secret key
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		as.logger.Printf("generateJWT[service]: Ошибка при подписании токена: %s", err)
		return "", err
	}

	as.logger.Debugf("generateJWT([service]: Токен успешно сгенерирован для пользователя: %s", user.Email)
	return tokenString, nil
}

// generatePasswordHash hashes user's password using bcrypt
func generatePasswordHash(password string) (string, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
