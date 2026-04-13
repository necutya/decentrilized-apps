package repo

import (
	"errors"

	"github.com/necutya/decentrilized_apps/lab3/node/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repo struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) ListEvents() ([]model.Event, error) {
	var events []model.Event
	return events, r.db.Find(&events).Error
}

func (r *Repo) GetEvent(id int64) (*model.Event, error) {
	var e model.Event
	if err := r.db.First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *Repo) DecrementSeats(eventID int64, seats int32) error {
	return r.db.Model(&model.Event{}).
		Where("id = ? AND available_seats >= ?", eventID, seats).
		Update("available_seats", gorm.Expr("available_seats - ?", seats)).Error
}

func (r *Repo) IncrementSeats(eventID int64, seats int32) error {
	return r.db.Model(&model.Event{}).
		Where("id = ?", eventID).
		Update("available_seats", gorm.Expr("available_seats + ?", seats)).Error
}

func (r *Repo) CreateBooking(b *model.Booking) error {
	return r.db.Create(b).Error
}

func (r *Repo) UpsertBooking(b *model.Booking) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"event_id", "user_id", "seats", "status", "total_price"}),
	}).Create(b).Error
}

func (r *Repo) GetBooking(id int64) (*model.Booking, error) {
	var b model.Booking
	if err := r.db.First(&b, id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *Repo) UpdateBookingStatus(id int64, status string) error {
	return r.db.Model(&model.Booking{}).Where("id = ?", id).Update("status", status).Error
}

func (r *Repo) ListBookingsByUser(userID string) ([]model.Booking, error) {
	var bookings []model.Booking
	return bookings, r.db.Where("user_id = ?", userID).Find(&bookings).Error
}

func (r *Repo) EventsEmpty() bool {
	var count int64
	r.db.Model(&model.Event{}).Count(&count)
	return count == 0
}

var ErrNotFound = errors.New("not found")

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
