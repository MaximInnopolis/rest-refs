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

var ErrReferralCodeNotActive = errors.New("реферальный код не существует")

// ReferralService represents service for handling referrals
type ReferralService struct {
	repo                repository.ReferralRepo
	logger              *logrus.Logger
	referralCodeService *ReferralCodeService
}

// NewReferralService creates new instance of ReferralService with repository, authService
func NewReferralService(repo repository.ReferralRepo, referralCodeService *ReferralCodeService, logger *logrus.Logger) *ReferralService {
	return &ReferralService{
		repo:                repo,
		referralCodeService: referralCodeService,
		logger:              logger,
	}
}

func (r *ReferralService) GetReferralsByReferrerID(referrerID int) ([]models.ReferralInfoResponse, error) {
	r.logger.Debugf("GetReferralsByReferrerID[service]: Получение рефералов для пользователя с id: %d", referrerID)

	referrals, err := r.repo.GetReferralsByReferrerID(referrerID)
	if err != nil {
		r.logger.Errorf("GetReferralsByReferrerID[service]: Ошибка при получении рефералов для пользователя с id: %d: %s", referrerID, err)
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

	r.logger.Infof("GetReferralsByReferrerID[service]: Рефералы успешно получены для пользователя с id: %d", referrerID)
	return response, nil
}

func (r *ReferralService) RegisterWithReferralCode(referralCode string, user models.User) error {
	r.logger.Debugf("RegisterWithReferralCode[service]: Регистрация реферала:"+
		" %s с реферальным кодом: %s", user.Email, referralCode)

	// Attempt to get active referral code
	codeID, err := r.referralCodeService.GetIDByReferralCode(referralCode)
	if err != nil {
		return err
	}
	if codeID == 0 {
		return ErrReferralCodeNotActive
	}

	// Attempt to create user using service
	err = r.referralCodeService.authService.RegisterUser(user)
	if err != nil {
		return err
	}

	referrerID, err := r.referralCodeService.GetReferrerIDByReferralCode(referralCode)
	if err != nil {
		return err
	}

	referral := models.Referral{
		Email:          user.Email,
		ReferralCodeID: codeID,
		ReferrerID:     referrerID,
	}

	// Save new referral in repository
	err = r.repo.Create(referral)
	if err != nil {
		r.logger.Errorf("RegisterWithReferralCode[service]: Ошибка при создании реферала в базе: %s", err)
		return err
	}

	r.logger.Infof("RegisterUser[service]: Реферал с email: %s успешно зарегистрирован", user.Email)
	return nil
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
