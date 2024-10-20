package api

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/sirupsen/logrus"
	"rest-refs/internal/app/models"
	"rest-refs/internal/app/repository"
)

var ErrReferralCodeAlreadyExists = errors.New("активный реферальный код уже существует")

const referralCodeLength = 8

// ReferralService represents service for handling referrals
type ReferralService struct {
	repo   repository.ReferralRepo
	logger *logrus.Logger
}

// NewReferralService creates new instance of ReferralService with repository
func NewReferralService(repo repository.ReferralRepo, logger *logrus.Logger) *ReferralService {
	return &ReferralService{
		repo:   repo,
		logger: logger,
	}
}

// CreateReferralCode creates new referral code using repository and returns created referral code
func (r *ReferralService) CreateReferralCode(referralCode models.ReferralCode) (models.ReferralCode, error) {
	r.logger.Debugf("CreateReferralCode[service]: Создание реферального кода пользователя %d", referralCode.ReferrerID)

	// Generate referral code
	code, err := generateReferralCode()
	if err != nil {
		return models.ReferralCode{}, err
	}

	referralCode.Code = code

	// Check if there's already active referral code for referrer
	_, err = r.repo.GetActiveReferralCodeByReferrerID(referralCode.ReferrerID)
	if err == nil {
		r.logger.Errorf("CreateReferralCode[service]: Создание реферального кода не удалось: "+
			"У пользователя с id %d уже есть активный реферальный код", referralCode.ReferrerID)
		return models.ReferralCode{}, ErrReferralCodeAlreadyExists
	}

	// If no active referral code, create new one
	createdCode, err := r.repo.Create(referralCode)
	if err != nil {
		r.logger.Errorf("CreateReferralCode[service]: Ошибка создания реферального кода в базе: %s", err)
		return models.ReferralCode{}, err
	}

	r.logger.Infof("CreateReferralCode[service]: Реферальный код создан")
	return createdCode, nil
}

func (r *ReferralService) DeleteReferralCode(referrerID int) error {
	r.logger.Debugf("DeleteReferralCode[service]: Удаление реферального пользователя с id %d", referrerID)

	// Check if there's already active referral code for referrer
	_, err := r.repo.GetActiveReferralCodeByReferrerID(referrerID)
	if err != nil {
		r.logger.Errorf("DeleteReferralCode[service]: Ошибка при получении активного кода: %s", err)
		return err
	}

	err = r.repo.Delete(referrerID)
	if err != nil {
		r.logger.Errorf("DeleteReferralCode[service]: Ошибка при удалении реферального кода: %s", err)
		return err
	}

	r.logger.Infof("DeleteReferralCode[service]: Реферальный код пользователя с ID %d успешно удален", referrerID)
	return nil
}

func (r *ReferralService) GetReferralCodeByReferrerEmail(email string) (models.ReferralCode, error) {
	r.logger.Debugf("GetReferralCodeByReferrerEmail[service]: Получение реферального кода для email: %s", email)

	userID, err := r.repo.GetReferrerIDByEmail(email)
	if err != nil {
		r.logger.Errorf("GetReferralCodeByReferrerEmail[service]: Ошибка при получении id пользователя по email %s: %s", email, err)
		return models.ReferralCode{}, err
	}

	code, err := r.repo.GetActiveReferralCodeByReferrerID(userID)
	if err != nil {
		r.logger.Errorf("GetReferralCodeByReferrerEmail[service]: Ошибка при получении активного реферального кода по email %s: %s", email, err)
		return models.ReferralCode{}, err
	}

	r.logger.Infof("GetReferralCodeByReferrerEmail[service]: Реферальный код успешно получен для email: %s", email)
	return code, nil
}

func (r *ReferralService) GetReferralsByReferrerID(referrerID int) ([]models.ReferralInfoResponse, error) {
	r.logger.Debugf("GetReferralsByReferrerID[service]: Получение рефералов для пользователя с id %d", referrerID)

	referrals, err := r.repo.GetReferralsByReferrerID(referrerID)
	if err != nil {
		r.logger.Errorf("GetReferralsByReferrerID[service]: Ошибка при получении рефералов для пользователя с id %d: %s", referrerID, err)
		return nil, err
	}

	var response []models.ReferralInfoResponse
	for _, referral := range referrals {
		response = append(response, models.ReferralInfoResponse{
			ReferrerID: referral.ReferrerID,
			Email:      referral.Email,
			CreatedAt:  referral.CreatedAt,
		})
	}

	r.logger.Infof("GetReferralsByReferrerID[service]: Рефералы успешно получены для пользователя с id %d", referrerID)
	return response, nil
}

func (r *ReferralService) RegisterWithReferralCode(referralCode string, user models.User) error {}

// generateReferralCode generates unique secure referral code
func generateReferralCode() (string, error) {
	randomBytes := make([]byte, referralCodeLength)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	code := base64.RawURLEncoding.EncodeToString(randomBytes)

	return strings.ToUpper(code[:referralCodeLength]), nil
}
