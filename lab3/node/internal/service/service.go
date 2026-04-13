package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/necutya/decentrilized_apps/lab3/node/gen/nodepb"
	"github.com/necutya/decentrilized_apps/lab3/node/internal/model"
	"github.com/necutya/decentrilized_apps/lab3/node/internal/p2p"
	"github.com/necutya/decentrilized_apps/lab3/node/internal/repo"
)

type Service struct {
	repo        *repo.Repo
	broadcaster *p2p.Broadcaster
}

func New(r *repo.Repo, b *p2p.Broadcaster) *Service {
	return &Service{repo: r, broadcaster: b}
}

func (s *Service) ListEvents() ([]model.Event, error) {
	return s.repo.ListEvents()
}

func (s *Service) BookTicket(eventID int64, seats int32, userID string) (*model.Booking, error) {
	event, err := s.repo.GetEvent(eventID)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, fmt.Errorf("event %d not found", eventID)
		}
		return nil, err
	}
	if event.AvailableSeats < seats {
		return nil, errors.New("not enough available seats")
	}

	if err := s.repo.DecrementSeats(eventID, seats); err != nil {
		return nil, err
	}

	booking := &model.Booking{
		EventID:    eventID,
		UserID:     userID,
		Seats:      seats,
		Status:     "confirmed",
		TotalPrice: float64(seats) * event.Price,
	}
	if err := s.repo.CreateBooking(booking); err != nil {
		return nil, err
	}

	payload, _ := json.Marshal(booking)
	s.broadcaster.Broadcast(nodepb.SyncAction_BOOK, payload)

	return booking, nil
}

func (s *Service) CancelBooking(id int64, userID string) error {
	booking, err := s.repo.GetBooking(id)
	if err != nil {
		if repo.IsNotFound(err) {
			return fmt.Errorf("booking %d not found", id)
		}
		return err
	}
	if booking.UserID != userID {
		return errors.New("unauthorized: booking belongs to another user")
	}
	if booking.Status == "cancelled" {
		return errors.New("booking already cancelled")
	}

	if err := s.repo.UpdateBookingStatus(id, "cancelled"); err != nil {
		return err
	}
	if err := s.repo.IncrementSeats(booking.EventID, booking.Seats); err != nil {
		return err
	}

	booking.Status = "cancelled"
	payload, _ := json.Marshal(booking)
	s.broadcaster.Broadcast(nodepb.SyncAction_CANCEL, payload)

	return nil
}

func (s *Service) ListBookings(userID string) ([]model.Booking, error) {
	return s.repo.ListBookingsByUser(userID)
}
