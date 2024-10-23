package api

import (
	"errors"

	"github.com/sirupsen/logrus"
	"rest-refs/internal/app/models"
	"rest-refs/internal/app/repository"
)

var ErrReferralCodeAlreadyExists = errors.New("активный реферальный код уже существует")

const referralCodeLength = 8

// ReferralCodeService represents service for handling referral codes
type ReferralCodeService struct {
	repo        repository.ReferralCodeRepo
	logger      *logrus.Logger
	authService *AuthService
}

// NewReferralCodeService creates new instance of ReferralCodeService with repository, authService
func NewReferralCodeService(repo repository.ReferralCodeRepo, authService *AuthService, logger *logrus.Logger) *ReferralCodeService {
	return &ReferralCodeService{
		repo:        repo,
		authService: authService,
		logger:      logger,
	}
}

// CreateReferralCode creates new referral code using repository and returns created referral code
func (r *ReferralCodeService) CreateReferralCode(referralCode models.ReferralCode) (models.ReferralCode, error) {
	r.logger.Debugf("Create[service]: Создание реферального кода пользователя c id: %d", referralCode.ReferrerID)

	// Generate referral code
	code, err := generateReferralCode()
	if err != nil {
		return models.ReferralCode{}, err
	}

	referralCode.Code = code

	// Check if there's already active referral code for referrer
	_, err = r.repo.GetActiveReferralCodeByUserID(referralCode.ReferrerID)
	if err == nil {
		r.logger.Errorf("Create[service]: Создание реферального кода не удалось: "+
			"У пользователя с id: %d уже есть активный реферальный код", referralCode.ReferrerID)
		return models.ReferralCode{}, ErrReferralCodeAlreadyExists
	}

	// If no active referral code, create new one
	createdCode, err := r.repo.Create(referralCode)
	if err != nil {
		r.logger.Errorf("Create[service]: Ошибка создания реферального кода для пользователя с id:"+
			" %d в базе: %s", referralCode.ReferrerID, err)
		return models.ReferralCode{}, err
	}

	r.logger.Infof("Create[service]: Реферальный код создан для пользователя с id: %d",
		referralCode.ReferrerID)
	return createdCode, nil
}

// DeleteReferralCode removes active referral code for specified referrer ID
func (r *ReferralCodeService) DeleteReferralCode(referrerID int) error {
	r.logger.Debugf("DeleteReferralCode[service]: Удаление реферального пользователя с id: %d", referrerID)

	// Check if there's already active referral code for referrer
	referralCode, err := r.repo.GetActiveReferralCodeByUserID(referrerID)
	if err != nil {
		r.logger.Errorf("DeleteReferralCode[service]: Ошибка при получении активного реферального кода"+
			" для пользователя с id: %d: %s", referrerID, err)
		return err
	}

	// Delete the active referral code by its ID
	err = r.repo.DeleteActiveReferralCodeByID(referralCode.ID)
	if err != nil {
		r.logger.Errorf("DeleteReferralCode[service]: Ошибка при удалении реферального кода для пользователя"+
			" с id: %d: %s", referrerID, err)
		return err
	}

	r.logger.Infof("DeleteReferralCode[service]: Реферальный код пользователя с id: %d успешно удален", referrerID)
	return nil
}

// GetReferralCodeByReferrerEmail retrieves the active referral code associated with a specific user's email
func (r *ReferralCodeService) GetReferralCodeByReferrerEmail(email string) (models.ReferralCode, error) {
	r.logger.Debugf("GetReferralCodeByReferrerEmail[service]: Получение реферального кода для email: %s", email)

	// Fetch user by email
	user, err := r.authService.GetUserByEmail(email)
	if err != nil {
		r.logger.Errorf("GetReferralCodeByReferrerEmail[service]: Ошибка при получении id пользователя"+
			" по email %s: %s", email, err)
		return models.ReferralCode{}, err
	}

	// Retrieve active referral code for the user
	code, err := r.repo.GetActiveReferralCodeByUserID(user.ID)
	if err != nil {
		r.logger.Errorf("GetReferralCodeByReferrerEmail[service]: Ошибка при получении активного реферального кода"+
			" по email %s: %s", email, err)
		return models.ReferralCode{}, err
	}

	r.logger.Infof("GetReferralCodeByReferrerEmail[service]: Реферальный код успешно получен для email: %s", email)
	return code, nil
}

// GetIDByReferralCode retrieves ID of a referral code from repository
func (r *ReferralCodeService) GetIDByReferralCode(code string) (int, error) {
	r.logger.Debugf("GetIDByReferralCode[service]: Получение id реферального кода: %s", code)
	return r.repo.GetIDByReferralCode(code)
}

// GetReferrerIDByReferralCode retrieves referrer ID associated with specific referral code
func (r *ReferralCodeService) GetReferrerIDByReferralCode(code string) (int, error) {
	r.logger.Debugf("GetReferrerIDByReferralCode[service]: Получение id реферера по реферальному коду: %s", code)
	return r.repo.GetReferrerIDByReferralCode(code)
}
