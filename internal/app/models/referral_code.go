package models

import "time"

type ReferralCode struct {
	ID         int        `json:"id"`
	Code       string     `json:"code"`
	Expiration time.Time  `json:"expires_at"`
	ReferrerID int        `json:"referrer_id"`
	Referrals  []Referral `json:"referrals,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
