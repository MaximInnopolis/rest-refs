package models

import (
	"time"
)

type Referral struct {
	ID             int       `json:"id"`
	ReferrerID     int       `json:"referrer_id"`
	ReferralCodeID int       `json:"referral_code_id"`
	ReferredUserID int       `json:"referred_user_id"`
	CreatedAt      time.Time `json:"created_at"`
}
