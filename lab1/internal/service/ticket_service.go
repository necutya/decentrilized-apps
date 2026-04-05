package service

import (
	"errors"

	"github.com/necutya/decentrilized_apps/lab1/internal/model"
	"github.com/necutya/decentrilized_apps/lab1/internal/repo"
)

type TicketService struct {
	events   *repo.EventRepo
	bookings *repo.BookingRepo
}

func NewTicketService(events *repo.EventRepo, bookings *repo.BookingRepo) *TicketService {
	return &TicketService{events: events, bookings: bookings}
}

func (s *TicketService) ListEvents() ([]model.Event, error) {
	return s.events.List()
}

func (s *TicketService) GetEvent(id int64) (*model.Event, error) {
	return s.events.FindByID(id)
}

func (s *TicketService) BookTicket(userID uint, eventID int64, seats int32) (*model.Booking, error) {
	event, err := s.events.FindByID(eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}
	if event.AvailableSeats < seats {
		return nil, errors.New("not enough seats available")
	}
	if err := s.events.DecrementSeats(eventID, seats); err != nil {
		return nil, errors.New("could not reserve seats")
	}
	b := &model.Booking{
		UserID:     userID,
		EventID:    eventID,
		Seats:      seats,
		Status:     "confirmed",
		TotalPrice: float64(seats) * event.Price,
	}
	if err := s.bookings.Create(b); err != nil {
		// roll back seat decrement
		_ = s.events.IncrementSeats(eventID, seats)
		return nil, err
	}
	b.Event = *event
	return b, nil
}

func (s *TicketService) ListMyBookings(userID uint) ([]model.Booking, error) {
	return s.bookings.ListByUser(userID)
}

func (s *TicketService) CancelBooking(userID uint, bookingID int64) error {
	b, err := s.bookings.FindByID(bookingID)
	if err != nil {
		return errors.New("booking not found")
	}
	if b.UserID != userID {
		return errors.New("not your booking")
	}
	if b.Status == "cancelled" {
		return errors.New("already cancelled")
	}
	if err := s.bookings.UpdateStatus(bookingID, "cancelled"); err != nil {
		return err
	}
	return s.events.IncrementSeats(b.EventID, b.Seats)
}
