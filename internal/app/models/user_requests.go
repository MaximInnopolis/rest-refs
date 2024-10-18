package models

type RegisterRequest struct {
	Email        string `json:"email" binding:"required"`
	Password     string `json:"password" binding:"required"`
	ReferralCode string `json:"referral_code,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}
