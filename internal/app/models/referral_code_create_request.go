package models

type ReferralCodeCreateRequest struct {
	ExpirationDate string `json:"expiration_date" binding:"required"`
}
