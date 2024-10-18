package models

import "time"

type ReferralInfoResponse struct {
	ReferredUserID uint      `json:"referred_user_id"`
	Email          string    `json:"email"`
	RegisteredAt   time.Time `json:"registered_at"`
}
