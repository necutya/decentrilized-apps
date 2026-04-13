package model

import "time"

type Event struct {
	ID             int64 `gorm:"primaryKey"`
	Title          string
	Venue          string
	Date           time.Time
	AvailableSeats int32
	TotalSeats     int32
	Price          float64
}

type Booking struct {
	ID         int64  `gorm:"primaryKey;autoIncrement"`
	EventID    int64
	UserID     string
	Seats      int32
	Status     string
	TotalPrice float64
}
