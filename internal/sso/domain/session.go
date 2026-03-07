package domain

import "time"

type Session struct {
	RefreshToken string
	DeviceName   string
	CreatedAt    time.Time
}
