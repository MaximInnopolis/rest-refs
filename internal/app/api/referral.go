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
	r.logger.Infof("CreateReferralCode[service]: Создание реферального кода пользователя %d", referralCode.ReferrerID)

	// Generate referral code
	code, err := generateReferralCode()
	if err != nil {
		return models.ReferralCode{}, err
	}

	referralCode.Code = code

	// Check if there's already active referral code for referrer
	existingCode, err := r.repo.GetActiveReferralCodeByReferrerID(referralCode.ReferrerID)
	if err != nil {
		r.logger.Errorf("CreateReferralCode[service]: Ошибка при получении id реферала из базы: %s", err)
		return models.ReferralCode{}, err
	}

	// If there's already active code, return ErrReferralCodeAlreadyExists
	if existingCode != nil {
		r.logger.Errorf("CreateReferralCode[service]: Реферальный код у реферера с id %d уже существует", referralCode.ReferrerID)
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
	r.logger.Infof("DeleteReferralCode[service]: Удаление реферального пользователя %d", referrerID)

	return r.repo.Delete(referrerID)
}

func (r *ReferralService) GetReferralCodeByEmail(email string) (models.ReferralCode, error) {}

func (r *ReferralService) RegisterWithReferralCode(referralCode string, user models.User) error {}

func (r *ReferralService) GetReferralsByReferrerID(referrerID int) ([]models.ReferralInfoResponse, error) {
}

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
