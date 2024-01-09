package models

import "time"

type UserInfo struct {
	Username string `json:"login"`
	Password string `json:"password"`
}

type Session struct {
	Username  string
	ExpiresAt time.Time
	Token     string
}

type Order struct {
	Number     string    `json:"number"`
	Username   string    `json:"-"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Balance struct {
	User      string  `json:"-"`
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type WithdrawalPost struct {
	Order string
	Sum   float64
}

type Withdrawal struct {
	User        string    `json:"-"`
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}
