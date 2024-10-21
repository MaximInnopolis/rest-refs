package models

import "time"

type ReferralInfoResponse struct {
	ReferralID int       `json:"referral_id"`
	ReferrerID int       `json:"referrer_id"`
	Email      string    `json:"email"`
	CreatedAt  time.Time `json:"created_at"`
}
