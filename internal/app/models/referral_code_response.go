package models

import "time"

type ReferralCodeResponse struct {
	Code       string    `json:"code"`
	Expiration time.Time `json:"expiration"`
}
