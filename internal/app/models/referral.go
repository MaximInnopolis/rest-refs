package models

import (
	"time"
)

type Referral struct {
	ID             int       `json:"id"`
	Email          string    `json:"email"`
	ReferralCodeID int       `json:"referral_code_id"`
	ReferrerID     int       `json:"referrer_id"`
	CreatedAt      time.Time `json:"created_at"`
}
