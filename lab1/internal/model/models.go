package model

import "time"

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"uniqueIndex;not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
}

type Event struct {
	ID             int64   `gorm:"primaryKey;autoIncrement"`
	Title          string  `gorm:"not null"`
	Venue          string  `gorm:"not null"`
	Date           time.Time
	TotalSeats     int32   `gorm:"not null"`
	AvailableSeats int32   `gorm:"not null"`
	Price          float64 `gorm:"not null"`
}

type Booking struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	UserID    uint   `gorm:"not null;index"`
	EventID   int64  `gorm:"not null;index"`
	Event     Event  `gorm:"foreignKey:EventID"`
	Seats     int32  `gorm:"not null"`
	Status    string `gorm:"not null;default:confirmed"` // confirmed | cancelled
	TotalPrice float64 `gorm:"not null"`
}
